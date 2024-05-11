package store

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/adwski/vidi/internal/session"
	"github.com/dgraph-io/ristretto"
	jsoniter "github.com/json-iterator/go"
	"github.com/redis/go-redis/v9"
)

// Ristretto params.
const (
	ristrettoNumCounters = 1e7     // number of keys to track frequency of (10M).
	ristrettoMaxCost     = 1 << 30 // maximum cost of cache (1GB).
	ristrettoBufferItems = 64      // number of keys per Get buffer.
	ristrettoDefaultCost = 1
)

var (
	ErrNotFound = errors.New("session not found")
)

// Store is a session store.
// It supports Name parameter (which basically is session key prefix like "<prefix>:<key>")
// Redis session is stored with preconfigured TTL
//
// Also it uses in-memory cache, which allows callers to check session very frequently
// without redis interaction. Just like redis, in-memory cache also has TTL and by design
// it is half of configured redis TTL. This helps to survive traffic bursts and still
// update session TTL in redis.
//
// Only GetExpireCached() supports session caching.
type Store struct {
	logger   *zap.Logger
	r        *redis.Client
	cache    *ristretto.Cache
	enc      jsoniter.API
	name     []byte
	ttl      time.Duration
	cacheTTL time.Duration
}

type Config struct {
	Logger   *zap.Logger
	Name     string
	RedisDSN string
	TTL      time.Duration
}

func NewStore(cfg *Config) (*Store, error) {
	u, err := url.Parse(cfg.RedisDSN)
	if err != nil {
		return nil, fmt.Errorf("cannot parse redis DSN: %w", err)
	}
	db, errDB := strconv.Atoi(strings.TrimPrefix(u.Path, "/"))
	if errDB != nil {
		return nil, fmt.Errorf("cannot parse redis DB number: %w", errDB)
	}
	cache, errRi := ristretto.NewCache(&ristretto.Config{
		NumCounters: ristrettoNumCounters,
		MaxCost:     ristrettoMaxCost, // basically size of cache if cost of each elem is 1
		BufferItems: ristrettoBufferItems,
	})
	if errRi != nil {
		return nil, fmt.Errorf("cannot initialize ristretto cache: %w", errRi)
	}
	pass, _ := u.User.Password() // empty if not ok
	r := redis.NewClient(&redis.Options{
		Addr:     u.Host,
		Username: u.User.Username(),
		Password: pass,
		DB:       db,
	})
	return &Store{
		logger: cfg.Logger.With(
			zap.String("component", "session-store"),
			zap.String("name", cfg.Name)),
		r:     r,
		cache: cache,
		enc:   jsoniter.ConfigCompatibleWithStandardLibrary,
		name:  []byte(cfg.Name),
		ttl:   cfg.TTL,

		// Local cache ttl is half of redis ttl.
		// Because of this after half-time we goto redis and update expiration,
		// since we don't want to lose session in redis.
		cacheTTL: cfg.TTL / 2, //nolint:mnd  // explained above, no need for constant
	}, nil
}

func (s *Store) Close() {
	if err := s.r.Close(); err != nil {
		s.logger.Error("error while closing redis connector", zap.Error(err))
	}
}

func (s *Store) getFullKey(key string) string {
	b := make([]byte, 0, len(s.name)+len(key)+1)
	return string(append(append(append(b, s.name...), ':'), []byte(key)...))
}

func (s *Store) Set(ctx context.Context, sess *session.Session) error {
	data, err := s.enc.Marshal(sess)
	if err != nil {
		return fmt.Errorf("cannot encode session: %w", err)
	}
	if err = s.r.Set(ctx, s.getFullKey(sess.ID), data, s.ttl).Err(); err != nil {
		return fmt.Errorf("cannot store session: %w", err)
	}
	return nil
}

func (s *Store) Get(ctx context.Context, key string) (*session.Session, error) {
	b, err := s.r.Get(ctx, s.getFullKey(key)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("cannot retrieve session: %w", err)
	}
	var sess session.Session
	if err = s.enc.Unmarshal(b, &sess); err != nil {
		return nil, fmt.Errorf("cannot decode session: %w", err)
	}
	return &sess, nil
}

func (s *Store) GetExpireCached(ctx context.Context, key string) (*session.Session, error) {
	value, found := s.cache.Get(key)
	if found {
		v, _ := value.(session.Session)
		return &v, nil
	}
	sess, err := s.GetExpire(ctx, key)
	if err != nil {
		return nil, err
	}
	// We do not care if value is actually set.
	// During continuous get calls it will be stored eventually.
	// TODO Find out is it better to store copies or pointers
	_ = s.cache.SetWithTTL(key, *sess, ristrettoDefaultCost, s.cacheTTL)
	return sess, err
}

func (s *Store) GetExpire(ctx context.Context, key string) (*session.Session, error) {
	fullKey := s.getFullKey(key)
	b, err := s.r.Get(ctx, fullKey).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("cannot retrieve session: %w", err)
	}
	err = s.r.Expire(ctx, fullKey, s.ttl).Err()
	if err != nil {
		return nil, fmt.Errorf("cannot update ttl: %w", err)
	}
	var sess session.Session
	if err = s.enc.Unmarshal(b, &sess); err != nil {
		return nil, fmt.Errorf("cannot decode session: %w", err)
	}
	return &sess, nil
}

func (s *Store) Delete(ctx context.Context, key string) error {
	if err := s.r.Del(ctx, s.getFullKey(key)).Err(); err != nil {
		if errors.Is(err, redis.Nil) {
			return nil // key might already be expired, and it is ok
		}
		return fmt.Errorf("cannot delete session: %w", err)
	}
	return nil
}

func (s *Store) Name() string {
	return string(s.name)
}

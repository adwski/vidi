package store

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/adwski/vidi/internal/session"
	jsoniter "github.com/json-iterator/go"
	"github.com/redis/go-redis/v9"
)

type Store struct {
	r    *redis.Client
	enc  jsoniter.API
	name []byte
	ttl  time.Duration
}

type Config struct {
	Name     string
	RedisDSN string
	TTL      time.Duration
}

func NewStore(cfg *Config) (*Store, error) {
	u, err := url.Parse(cfg.RedisDSN)
	if err != nil {
		return nil, fmt.Errorf("cannot parse redis DSN")
	}
	db, errDB := strconv.Atoi(strings.TrimPrefix(u.Path, "/"))
	if errDB != nil {
		return nil, fmt.Errorf("cannot parse redis DB number")
	}
	pass, _ := u.User.Password() // empty if not ok
	r := redis.NewClient(&redis.Options{
		Addr:     u.Host,
		Username: u.User.Username(),
		Password: pass,
		DB:       db,
	})
	return &Store{
		r:    r,
		enc:  jsoniter.ConfigCompatibleWithStandardLibrary,
		name: []byte(cfg.Name),
		ttl:  cfg.TTL,
	}, nil
}

func (s *Store) Close() error {
	if err := s.r.Close(); err != nil {
		return fmt.Errorf("error while closing redis connector: %w", err)
	}
	return nil
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
		return nil, fmt.Errorf("cannot retrieve session: %w", err)
	}
	var sess session.Session
	if err = s.enc.Unmarshal(b, &sess); err != nil {
		return nil, fmt.Errorf("cannot decode session: %w", err)
	}
	return &sess, nil
}

func (s *Store) Delete(ctx context.Context, key string) error {
	if err := s.r.Del(ctx, s.getFullKey(key)).Err(); err != nil {
		// TODO Check what happen if we delete expired key
		return fmt.Errorf("cannot delete session: %w", err)
	}
	return nil
}

func (s *Store) Name() string {
	return string(s.name)
}

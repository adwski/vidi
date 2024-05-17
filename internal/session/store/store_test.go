package store

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/adwski/vidi/internal/session"
	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestStore_Name(t *testing.T) {
	s := Store{name: []byte("test")}
	assert.Equal(t, "test", s.Name())
}

func TestNewStore(t *testing.T) {
	type args struct {
		cfg *Config
	}
	type want struct {
		errMsg string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "ok",
			args: args{
				cfg: &Config{
					Name:     "test",
					RedisDSN: "redis://localhost:1111/0",
					TTL:      300,
				},
			},
		},
		{
			name: "small ttl",
			args: args{
				cfg: &Config{
					Name:     "test",
					RedisDSN: "redis://localhost:1111/0",
					TTL:      1,
				},
			},
		},
		{
			name: "invalid db",
			args: args{
				cfg: &Config{
					Name:     "test",
					RedisDSN: "redis://localhost:1111/qweqweq",
					TTL:      300,
				},
			},
			want: want{
				errMsg: "cannot parse redis DB number",
			},
		},
		{
			name: "empty name",
			args: args{
				cfg: &Config{
					RedisDSN: "redis://localhost:1111/1",
					TTL:      300,
				},
			},
			want: want{
				errMsg: "store name is required",
			},
		},
		{
			name: "invalid dsn",
			args: args{
				cfg: &Config{
					Name:     "test",
					RedisDSN: "://asdsad",
					TTL:      300,
				},
			},
			want: want{
				errMsg: "cannot parse redis DSN",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := zap.NewDevelopment()
			require.NoError(t, err)

			tt.args.cfg.Logger = logger

			store, err := NewStore(tt.args.cfg)
			if tt.want.errMsg != "" {
				assert.Contains(t, err.Error(), tt.want.errMsg)
				assert.Nil(t, store)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, store)

				assert.NotNil(t, store.cache)
				assert.NotNil(t, store.r)
				assert.NotNil(t, store.enc)
				assert.GreaterOrEqual(t, minTTL, store.ttl)
				assert.Equal(t, store.cacheTTL, store.ttl/2)
				store.Close()
			}
		})
	}
}

func TestStore_NewClose(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	store, err := NewStore(&Config{
		Logger:   logger,
		Name:     "test",
		RedisDSN: "redis://localhost:1111/0",
		TTL:      300,
	})
	require.NoError(t, err)
	store.Close()
	store.Close()
}

func TestStore_SetDelete(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	name := "test"
	id := "testsession"
	key := name + ":" + id
	ttl := 600 * time.Second

	store, err := NewStore(&Config{
		Logger:   logger,
		Name:     "test",
		RedisDSN: "redis://placeholder:1111/0",
		TTL:      ttl,
	})
	require.NoError(t, err)
	require.NotNil(t, store)

	db, mock := redismock.NewClientMock()
	store.r = db

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mock.ExpectDel(key).SetVal(1)
	var delErr = errors.New("some error")
	mock.ExpectDel(key).SetErr(delErr)
	mock.ExpectDel(key).RedisNil()

	// delete session
	err = store.Delete(ctx, id)
	require.NoError(t, err)

	// delete err
	err = store.Delete(ctx, id)
	require.ErrorIs(t, err, delErr)

	// delete not existing
	err = store.Delete(ctx, id)
	require.NoError(t, err)

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestStore_SetGet(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	name := "test"
	id := "testsession"
	key := name + ":" + id
	ttl := 600 * time.Second

	store, err := NewStore(&Config{
		Logger:   logger,
		Name:     "test",
		RedisDSN: "redis://placeholder:1111/0",
		TTL:      ttl,
	})
	require.NoError(t, err)
	require.NotNil(t, store)

	db, mock := redismock.NewClientMock()
	store.r = db

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sess := &session.Session{
		ID:       id,
		VideoID:  "testvid",
		Location: "testloc",
		PartSize: 1000000,
	}
	b, err := json.Marshal(sess)
	require.NoError(t, err)

	// Set expectations
	mock.ExpectSet(key, b, ttl).SetVal("???") // idk what SetVal needs as value
	var setErr = errors.New("some error")
	mock.ExpectSet(key, b, ttl).SetErr(setErr)
	mock.ExpectGet(key).SetVal(string(b))

	// Set session
	err = store.Set(ctx, sess)
	require.NoError(t, err)

	err = store.Set(ctx, sess)
	require.ErrorIs(t, err, setErr)

	// get session
	sess2, err := store.Get(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, sess, sess2)

	// get invalid data
	mock.ExpectGet(key).SetVal("eqweqweqwe")
	sess3, err := store.Get(ctx, id)
	require.Nil(t, sess3)
	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot decode session")

	// get not existing key
	mock.ExpectGet(key).RedisNil()
	sess4, err := store.Get(ctx, id)
	require.Nil(t, sess4)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrNotFound)

	// get unknown error
	someErr := errors.New("some error")
	mock.ExpectGet(key).SetErr(someErr)
	sess5, err := store.Get(ctx, id)
	require.Nil(t, sess5)
	require.Error(t, err)
	require.ErrorIs(t, err, someErr)

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestStore_SetGetExpired(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	name := "test"
	id := "testsession"
	key := name + ":" + id
	ttl := 600 * time.Second

	store, err := NewStore(&Config{
		Logger:   logger,
		Name:     "test",
		RedisDSN: "redis://placeholder:1111/0",
		TTL:      ttl,
	})
	require.NoError(t, err)
	require.NotNil(t, store)

	db, mock := redismock.NewClientMock()
	store.r = db

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sess := &session.Session{
		ID:       id,
		VideoID:  "testvid",
		Location: "testloc",
		PartSize: 1000000,
	}
	b, err := json.Marshal(sess)
	require.NoError(t, err)

	// Set expectations
	mock.ExpectSet(key, b, ttl).SetVal("???")
	mock.ExpectGet(key).SetVal(string(b))
	mock.ExpectExpire(key, ttl).SetVal(true)

	// Set session
	err = store.Set(ctx, sess)
	require.NoError(t, err)

	// get session several times
	sess2, err := store.GetExpireCached(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, sess, sess2)

	time.Sleep(time.Second) // delay for ristretto to catch up

	// retrieve session 100 times
	for i := 0; i < 100; i++ {
		sess3, errG := store.GetExpireCached(ctx, id)
		require.NoError(t, errG)
		assert.Equal(t, sess, sess3)
	}

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestStore_SetGetExpiredErrors(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	name := "test"
	id := "testsession"
	key := name + ":" + id
	ttl := 600 * time.Second

	srcSess := &session.Session{
		ID:       id,
		VideoID:  "testvid",
		Location: "testloc",
		PartSize: 1000000,
	}
	b, err := json.Marshal(srcSess)
	require.NoError(t, err)

	store, err := NewStore(&Config{
		Logger:   logger,
		Name:     "test",
		RedisDSN: "redis://placeholder:1111/0",
		TTL:      ttl,
	})
	require.NoError(t, err)
	require.NotNil(t, store)

	db, mock := redismock.NewClientMock()
	store.r = db

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set expectations
	mock.ExpectGet(key).RedisNil()
	var getErr = errors.New("test")

	mock.ExpectGet(key).SetErr(getErr)

	mock.ExpectGet(key).SetVal(string(b))
	var expErr = errors.New("exp")
	mock.ExpectExpire(key, ttl).SetErr(expErr)

	mock.ExpectGet(key).SetVal("qweqweqweq")
	mock.ExpectExpire(key, ttl).SetVal(true)

	// get not existing session
	sess, err := store.GetExpireCached(ctx, id)
	require.ErrorIs(t, err, ErrNotFound)
	assert.Nil(t, sess)

	sess, err = store.GetExpireCached(ctx, id)
	require.ErrorIs(t, err, getErr)
	assert.Nil(t, sess)

	sess, err = store.GetExpireCached(ctx, id)
	require.ErrorIs(t, err, expErr)
	assert.Nil(t, sess)

	sess, err = store.GetExpireCached(ctx, id)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "cannot decode session")
	assert.Nil(t, sess)

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

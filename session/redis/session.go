package redis

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"github.com/JrMarcco/easy-web/session"
	"github.com/redis/go-redis/v9"
	"time"
)

//go:embed get_cmd.lua
var getRedisCmdLua string

//go:embed generate_cmd.lua
var generateRedisCmdLua string

var (
	errSessionNotFound = errors.New("[redis-session] session not found")
)

// RStore redis implementation for session store.
type RStore struct {
	client     redis.Cmdable
	prefix     string
	expiration time.Duration
}

func (r *RStore) Generate(ctx context.Context, id string) (session.Session, error) {
	key := r.key(id)
	_, err := r.client.Eval(
		ctx, generateRedisCmdLua, []string{key}, "_session_id", id, r.expiration.Milliseconds(),
	).Result()
	if err != nil {
		return nil, err
	}

	return &RSession{
		id:     id,
		key:    key,
		client: r.client,
	}, nil
}

func (r *RStore) Refresh(ctx context.Context, id string) error {
	key := r.key(id)
	res, err := r.client.Expire(ctx, key, r.expiration).Result()
	if err != nil {
		return err
	}

	if !res {
		return errSessionNotFound
	}
	return nil
}

func (r *RStore) Del(ctx context.Context, id string) error {
	_, err := r.client.Del(ctx, r.key(id)).Result()
	return err
}

func (r *RStore) Get(ctx context.Context, id string) (session.Session, error) {
	key := r.key(id)
	res, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	if res < 0 {
		return nil, errSessionNotFound
	}

	return &RSession{
		id:     id,
		key:    key,
		client: r.client,
	}, nil
}

func (r *RStore) key(id string) string {
	return fmt.Sprintf("%s:%s", r.prefix, id)
}

type RStoreOpt func(*RStore)

func RStoreWithPrefix(prefix string) RStoreOpt {
	return func(r *RStore) {
		r.prefix = prefix
	}
}

func RStoreWithExpiration(expiration time.Duration) RStoreOpt {
	return func(r *RStore) {
		r.expiration = expiration
	}
}

func NewRStore(client redis.Cmdable, opts ...RStoreOpt) *RStore {
	rs := &RStore{
		client:     client,
		prefix:     "session",
		expiration: time.Minute * 30,
	}

	for _, opt := range opts {
		opt(rs)
	}
	return rs
}

// RSession redis implementation for session.
type RSession struct {
	id     string
	key    string
	client redis.Cmdable
}

func (r *RSession) Get(ctx context.Context, key string) (any, error) {
	return r.client.HGet(ctx, r.key, key).Result()
}

func (r *RSession) Set(ctx context.Context, key string, value any) error {
	res, err := r.client.Eval(ctx, getRedisCmdLua, []string{r.key}, key, value).Int()
	if err != nil {
		return err
	}

	if res < 0 {
		return errSessionNotFound
	}
	return nil
}

func (r *RSession) Id() string {
	return r.id
}

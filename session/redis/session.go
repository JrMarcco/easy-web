package redis

import (
	"context"
	"errors"
	"github.com/jrmarcco/go-learning/web/session"
	"github.com/redis/go-redis/v9"
	"time"
)

var SessionNotFoundErr = errors.New("session is not found")

type Store struct {
	prefix     string
	client     redis.Cmdable
	expiration time.Duration
}

type StoreOpt func(store *Store)

func NewStore(client redis.Cmdable, opts ...StoreOpt) *Store {

	store := &Store{
		prefix: "_session_",
		client: client,
	}

	for _, opt := range opts {
		opt(store)
	}

	return store
}

func StoreWithPrefix(prefix string) StoreOpt {
	return func(store *Store) {
		store.prefix = prefix
	}
}

func (s *Store) Gen(ctx context.Context, id string) (session.Session, error) {
	redisKey := s.prefix + id
	_, err := s.client.HSet(ctx, redisKey, id, id).Result()
	if err != nil {
		return nil, err
	}

	_, err = s.client.Expire(ctx, redisKey, s.expiration).Result()
	if err != nil {
		return nil, err
	}

	return &Session{
		id:       id,
		redisKey: redisKey,
		client:   s.client,
	}, nil
}

func (s *Store) Refresh(ctx context.Context, id string) error {
	redisKey := s.prefix + id
	ok, err := s.client.Expire(ctx, redisKey, s.expiration).Result()
	if err != nil {
		return err
	}

	if !ok {
		return SessionNotFoundErr
	}

	return nil
}

func (s *Store) Del(ctx context.Context, id string) error {
	redisKey := s.prefix + id
	_, err := s.client.Del(ctx, redisKey).Result()
	return err
}

func (s *Store) Get(ctx context.Context, id string) (session.Session, error) {
	redisKey := s.prefix + id
	cnt, err := s.client.Exists(ctx, redisKey).Result()
	if err != nil {
		return nil, err
	}

	if cnt != 1 {
		return nil, SessionNotFoundErr
	}

	return &Session{
		id:       id,
		redisKey: redisKey,
		client:   s.client,
	}, nil
}

type Session struct {
	id       string
	redisKey string
	client   redis.Cmdable
}

func (s *Session) Get(ctx context.Context, key string) (any, error) {
	return s.client.HGet(ctx, s.redisKey, key).Result()
}

var setValLua = redis.NewScript(`
if redis.call("exists", KEYS[1])
then
	return redis.call("hset", KEYS[1], ARGV[1], ARGV[2])
else
	return -1
end
`)

func (s *Session) Set(ctx context.Context, key string, val any) error {
	keys := []string{s.redisKey}
	vals := []any{key, val}
	res, err := setValLua.Run(ctx, s.client, keys, vals).Int()
	if err != nil {
		return err
	}

	if res < 0 {
		return SessionNotFoundErr
	}

	return nil
}

func (s *Session) Id() string {
	return s.id
}

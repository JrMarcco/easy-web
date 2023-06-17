package memory

import (
	"context"
	"errors"
	"github.com/jrmarcco/go-learning/web/session"
	"github.com/patrickmn/go-cache"
	"sync"
	"time"
)

var (
	SessionNotFoundErr = errors.New("session is not found")
	KeyNotFoundErr     = errors.New("session key is not found")
)

type Store struct {
	sync.RWMutex

	sessions   *cache.Cache
	expiration time.Duration
}

func NewStore(expiration time.Duration) *Store {
	return &Store{
		sessions:   cache.New(expiration, time.Second),
		expiration: expiration,
	}
}

func (s *Store) Gen(_ context.Context, id string) (session.Session, error) {
	s.Lock()
	defer s.Unlock()

	sess := &Session{
		id: id,
	}

	s.sessions.Set(id, sess, s.expiration)
	return sess, nil
}

func (s *Store) Refresh(_ context.Context, id string) error {
	s.Lock()
	defer s.Unlock()

	val, ok := s.sessions.Get(id)
	if !ok {
		return SessionNotFoundErr
	}

	s.sessions.Set(id, val, s.expiration)
	return nil
}

func (s *Store) Del(_ context.Context, id string) error {
	s.Lock()
	defer s.Unlock()

	s.sessions.Delete(id)
	return nil
}

func (s *Store) Get(_ context.Context, id string) (session.Session, error) {
	s.RLock()
	defer s.RUnlock()

	val, ok := s.sessions.Get(id)
	if !ok {
		return nil, SessionNotFoundErr
	}

	return val.(*Session), nil
}

type Session struct {
	id   string
	vals sync.Map
}

func (s *Session) Get(_ context.Context, key string) (any, error) {
	val, ok := s.vals.Load(key)
	if !ok {
		return "", KeyNotFoundErr
	}
	return val, nil
}

func (s *Session) Set(_ context.Context, key string, val any) error {
	s.vals.Store(key, val)
	return nil
}

func (s *Session) Id() string {
	return s.id
}

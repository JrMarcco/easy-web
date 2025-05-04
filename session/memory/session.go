package memory

import (
	"context"
	"errors"
	"fmt"
	"github.com/JrMarcco/easy-web/session"
	"github.com/patrickmn/go-cache"
	"sync"
	"time"
)

var (
	errSessionNotFound    = errors.New("[mem-session] session not found")
	errSessionKeyNotFound = errors.New("[mem-session] session key not found")
)

// MemStore memory implementation for session store.
type MemStore struct {
	mu sync.RWMutex
	c  *cache.Cache
}

func (m *MemStore) Generate(_ context.Context, id string) (session.Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	s := &MemSession{
		id:   id,
		data: make(map[string]any),
	}

	m.c.SetDefault(id, s)
	return s, nil
}

func (m *MemStore) Refresh(_ context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if s, ok := m.c.Get(id); ok {
		m.c.SetDefault(id, s)
	}

	return errSessionNotFound
}

func (m *MemStore) Del(_ context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.c.Delete(id)
	return nil
}

func (m *MemStore) Get(_ context.Context, id string) (session.Session, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if s, ok := m.c.Get(id); ok {
		return s.(*MemSession), nil
	}

	return nil, errSessionNotFound
}

type MemStoreOpt func(*MemStore)

func MemStoreWithExpiration(expiration, cleanupInterval time.Duration) MemStoreOpt {
	return func(m *MemStore) {
		m.c = cache.New(expiration, cleanupInterval)
	}
}

func NewMemStore(opts ...MemStoreOpt) *MemStore {
	ms := &MemStore{
		c: cache.New(time.Minute*30, time.Second),
	}

	for _, opt := range opts {
		opt(ms)
	}
	return ms
}

// MemSession memory implementation for session memory.
type MemSession struct {
	mu   sync.RWMutex
	id   string
	data map[string]any
}

func (m *MemSession) Get(_ context.Context, key string) (any, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if val, ok := m.data[key]; ok {
		return val, nil
	}

	return "", fmt.Errorf("%w: %s", errSessionKeyNotFound, key)
}

func (m *MemSession) Set(_ context.Context, key string, value any) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.data[key] = value
	return nil
}

func (m *MemSession) Id() string {
	return m.id
}

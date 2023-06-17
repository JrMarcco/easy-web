package session

import (
	"github.com/google/uuid"
	"github.com/jrmarcco/go-learning/web"
	"github.com/jrmarcco/go-learning/web/session/memory"
	"time"
)

type Manager struct {
	store    Store
	provider Provider

	sessionKey string
	idFunc     func() string
}

type ManagerOpt func(manager *Manager)

func NewManager(opts ...ManagerOpt) *Manager {
	manager := &Manager{
		store:      memory.NewStore(30 * time.Minute),
		sessionKey: "_session_id",
		idFunc: func() string {
			return uuid.New().String()
		},
	}

	for _, opt := range opts {
		opt(manager)
	}
	return manager
}

func ManagerWithKey(key string) ManagerOpt {
	return func(manager *Manager) {
		manager.sessionKey = key
	}
}

func ManagerWithIdFunc(idFunc func() string) ManagerOpt {
	return func(manager *Manager) {
		manager.idFunc = idFunc
	}
}

func ManagerWithStore(store Store) ManagerOpt {
	return func(manager *Manager) {
		manager.store = store
	}
}

func ManagerWithProvider(provider Provider) ManagerOpt {
	return func(manager *Manager) {
		manager.provider = provider
	}
}

func (m *Manager) GetSession(ctx *web.Context) (Session, error) {

	if ctx.UserVals == nil {
		ctx.UserVals = make(map[string]any, 1)
	}

	val, ok := ctx.UserVals[m.sessionKey]
	if ok {
		return val.(Session), nil
	}

	sessionId, err := m.provider.Extract(ctx.Req)
	if err != nil {
		return nil, err
	}

	s, err := m.store.Get(ctx.Req.Context(), sessionId)
	if err != nil {
		return nil, err
	}

	ctx.UserVals[m.sessionKey] = s

	return s, err
}

func (m *Manager) InitSession(ctx *web.Context) (Session, error) {

	id := m.idFunc()
	s, err := m.store.Gen(ctx.Req.Context(), id)
	if err != nil {
		return nil, err
	}

	err = m.provider.Inject(id, ctx.Rsp)
	return s, err
}

func (m *Manager) RefreshSession(ctx *web.Context) error {
	session, err := m.GetSession(ctx)
	if err != nil {
		return err
	}

	return m.store.Refresh(ctx.Req.Context(), session.Id())
}

func (m *Manager) RemoveSession(ctx *web.Context) error {
	s, err := m.GetSession(ctx)
	if err != nil {
		return err
	}

	return m.store.Del(ctx.Req.Context(), s.Id())
}

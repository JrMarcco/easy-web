package session

import (
	easyweb "github.com/JrMarcco/easy-web"
	"github.com/google/uuid"
)

type Manager struct {
	Propagator
	Store

	sessionFieldName string
}

type ManagerOpt func(*Manager)

func WithSessionFieldName(sessionFieldName string) ManagerOpt {
	return func(m *Manager) {
		m.sessionFieldName = sessionFieldName
	}
}

func NewManager(store Store, propagator Propagator) *Manager {
	return &Manager{
		Store:            store,
		Propagator:       propagator,
		sessionFieldName: "_session",
	}
}

func (m *Manager) GetSession(ctx *easyweb.Context) (Session, error) {
	if ctx.UserValues == nil {
		ctx.UserValues = make(map[string]any, 1)
	}

	val, ok := ctx.UserValues[m.sessionFieldName]
	if ok {
		return val.(Session), nil
	}

	sId, err := m.Extract(ctx.Req)
	if err != nil {
		return nil, err
	}

	s, err := m.Get(ctx.Req.Context(), sId)
	if err != nil {
		return nil, err
	}

	ctx.UserValues[m.sessionFieldName] = s
	return s, nil
}

func (m *Manager) NewSession(ctx *easyweb.Context) (Session, error) {
	id := uuid.New().String()
	s, err := m.Generate(ctx.Req.Context(), id)
	if err != nil {
		return nil, err
	}

	err = m.Inject(id, ctx.Resp)

	if ctx.UserValues == nil {
		ctx.UserValues = make(map[string]any, 1)
	}
	ctx.UserValues[m.sessionFieldName] = s
	return s, err
}

func (m *Manager) RefreshSession(ctx *easyweb.Context) error {
	s, err := m.GetSession(ctx)
	if err != nil {
		return err
	}

	return m.Refresh(ctx.Req.Context(), s.Id())
}

func (m *Manager) DelSession(ctx *easyweb.Context) error {
	s, err := m.GetSession(ctx)
	if err != nil {
		return err
	}

	err = m.Store.Del(ctx.Req.Context(), s.Id())
	if err != nil {
		return err
	}

	return m.Propagator.Del(ctx.Resp)
}

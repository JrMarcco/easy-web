package session

import (
	"context"
	"net/http"
)

type Store interface {
	Generate(ctx context.Context, id string) (Session, error)
	Refresh(ctx context.Context, id string) error
	Del(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (Session, error)
}

type Session interface {
	Get(ctx context.Context, key string) (any, error)
	Set(ctx context.Context, key string, value any) error
	Id() string
}

type Propagator interface {
	Inject(id string, w http.ResponseWriter) error
	Extract(req *http.Request) (string, error)
	Del(w http.ResponseWriter) error
}

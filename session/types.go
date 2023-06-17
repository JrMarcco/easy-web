package session

import (
	"context"
	"net/http"
)

// Session 负责对 session 内的数据进行管理。
type Session interface {
	Get(ctx context.Context, key string) (any, error)
	Set(ctx context.Context, key string, val any) error
	Id() string
}

// Store 负责管理 session，
// 也就是说负责 Session 这个对象的各种操作。
// 即生成、刷新、删除、获取 Session。
type Store interface {
	Gen(ctx context.Context, id string) (Session, error)
	Refresh(ctx context.Context, id string) error
	Del(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (Session, error)
}

// Provider 负责与 http 请求交互，
// 进行 session id 与 http 响应的绑定，
// 以及从 http 请求提取 session id 等操作。
type Provider interface {
	Inject(id string, writer http.ResponseWriter) error
	Extract(req *http.Request) (string, error)
	Remove(writer http.ResponseWriter) error
}

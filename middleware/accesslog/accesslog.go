package accesslog

import (
	"encoding/json"
	"fmt"

	easyweb "github.com/JrMarcco/easy-web"
)

type MiddlewareBuilder struct {
	logFunc func(msg string)
}

func NewMwBuilder() *MiddlewareBuilder {
	return &MiddlewareBuilder{
		logFunc: func(msg string) {
			fmt.Println(msg)
		},
	}
}

func (b *MiddlewareBuilder) WithLogFunc(logFunc func(msg string)) *MiddlewareBuilder {
	b.logFunc = logFunc
	return b
}

func (b *MiddlewareBuilder) Build() easyweb.Middleware {
	return func(next easyweb.HandleFunc) easyweb.HandleFunc {
		return func(ctx *easyweb.Context) {
			defer func() {
				al := &accessLog{
					Host:   ctx.Req.Host,
					Method: ctx.Req.Method,
					Path:   ctx.Req.URL.Path,
					Route:  ctx.MatchedRoute,
					Status: ctx.StatusCode,
				}

				data, err := json.Marshal(al)
				if err != nil {
					b.logFunc(fmt.Sprintf("access log marshal error: %v", err))
				}

				b.logFunc(string(data))
			}()
			next(ctx)
		}
	}
}

type accessLog struct {
	Host   string `json:"host,omitempty"`
	Method string `json:"method,omitempty"`
	Path   string `json:"path,omitempty"`
	Route  string `json:"route,omitempty"`
	Status int    `json:"status,omitempty"`
}

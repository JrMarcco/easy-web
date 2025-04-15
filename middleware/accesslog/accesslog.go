package accesslog

import (
	"encoding/json"
	"fmt"

	"github.com/JrMarcco/easy_web"
)

type MwBuilder struct {
	logFunc func(msg string)
}

func NewMwBuilder() *MwBuilder {
	return &MwBuilder{
		logFunc: func(msg string) {
			fmt.Println(msg)
		},
	}
}

func (b *MwBuilder) WithLogFunc(logFunc func(msg string)) *MwBuilder {
	b.logFunc = logFunc
	return b
}

func (b *MwBuilder) Build() easy_web.MwFunc {
	return func(next easy_web.HdlFunc) easy_web.HdlFunc {
		return func(ctx *easy_web.Context) {
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

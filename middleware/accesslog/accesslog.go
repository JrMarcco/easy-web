package accesslog

import (
	"encoding/json"
	"github.com/jrmarcco/go-learning/web"
	"log"
)

type MiddleWareBuilder struct {
	logFunc func(msg string)
}

func NewBuilder() *MiddleWareBuilder {
	return &MiddleWareBuilder{
		logFunc: func(msg string) {
			log.Println(msg)
		},
	}
}
func (m *MiddleWareBuilder) LogFunc(logFunc func(msg string)) *MiddleWareBuilder {
	m.logFunc = logFunc
	return m
}

type logMsg struct {
	Host   string `json:"host,omitempty"`
	Method string `json:"method,omitempty"`
	Path   string `json:"path,omitempty"`
}

func (m *MiddleWareBuilder) Build() web.Middleware {
	return func(next web.HandleFunc) web.HandleFunc {
		return func(ctx *web.Context) {
			defer func() {
				lm := logMsg{
					Host:   ctx.Req.Host,
					Method: ctx.Req.Method,
					Path:   ctx.Req.URL.Path,
				}

				val, err := json.Marshal(lm)
				if err != nil {
					log.Println("[AccessLog] json marshal error", err)
					return
				}

				m.logFunc(string(val))
			}()
			next(ctx)
		}
	}
}

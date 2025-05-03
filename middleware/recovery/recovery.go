package recovery

import (
	easyweb "github.com/JrMarcco/easy-web"
	"log"
)

// MiddlewareBuilder should be the most outer in a middleware chain.
type MiddlewareBuilder struct {
	statusCode int
	errMsg     string
	logFunc    func(ctx *easyweb.Context)
}

// WithStatusCode the code returns to the front end when panicked.
// defaults to 500.
func (b *MiddlewareBuilder) WithStatusCode(statusCode int) *MiddlewareBuilder {
	b.statusCode = statusCode
	return b
}

// WithErrMsg the error message returns to the front end when panicked.
// defaults to "Internal Error"
func (b *MiddlewareBuilder) WithErrMsg(errMsg string) *MiddlewareBuilder {
	b.errMsg = errMsg
	return b
}

func (b *MiddlewareBuilder) WithLogFunc(logFunc func(ctx *easyweb.Context)) *MiddlewareBuilder {
	b.logFunc = logFunc
	return b
}

func (b *MiddlewareBuilder) Build() easyweb.Middleware {
	return func(next easyweb.HandleFunc) easyweb.HandleFunc {
		return func(ctx *easyweb.Context) {
			defer func() {
				if err := recover(); err != nil {
					ctx.StatusCode = b.statusCode
					ctx.Data = []byte(b.errMsg)

					b.logFunc(ctx)
				}
			}()
			next(ctx)
		}
	}
}

func NewMiddlewareBuilder() *MiddlewareBuilder {
	return &MiddlewareBuilder{
		statusCode: 500,
		errMsg:     "Internal Error",
		logFunc: func(ctx *easyweb.Context) {
			log.Printf("panic in path: %s", ctx.Req.URL.Path)
		},
	}
}

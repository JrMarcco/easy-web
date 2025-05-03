//go:build e2e

// for endpoint to endpoint test

package easyweb

import (
	"fmt"
	"net/http"
	"testing"
)

func TestServer_Middleware(t *testing.T) {
	svr := NewHttpServer()

	mwChain := MiddlewareChain{
		func(next HandleFunc) HandleFunc {
			return func(ctx *Context) {
				fmt.Println("middleware 1")
				next(ctx)
			}
		},
		func(next HandleFunc) HandleFunc {
			return func(ctx *Context) {
				fmt.Println("middleware 2")
				next(ctx)
			}
		},
		func(next HandleFunc) HandleFunc {
			return func(ctx *Context) {
				fmt.Println("middleware 3")
				next(ctx)
			}
		},
	}

	svr.Route(http.MethodGet, "/", func(ctx *Context) {
		fmt.Println("handler")
	}, mwChain...)

	err := svr.Start()
	if err != nil {
		t.Fatal(err)
	}
}

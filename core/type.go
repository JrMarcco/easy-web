package core

import "net/http"

type HdlFunc func(ctx *Context)

type Server interface {
	http.Handler

	Start(addr string) error
	RouteRegister(method string, path string, hdl HdlFunc)
}

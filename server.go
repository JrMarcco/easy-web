package easy_web

import (
	"log"
	"net"
	"net/http"
)

var _ Server = (*HttpSvr)(nil)

// HdlFunc is a handler function for a route
type HdlFunc func(ctx *Context)

// MwFunc is a middleware function that returns a new HdlFunc
// It is an aop implementation.
type MwFunc func(next HdlFunc) HdlFunc

// MwChain is a chain of middleware functions
type MwChain []MwFunc

type Server interface {
	http.Handler

	Start() error
	RouteRegister(method string, path string, hdl HdlFunc, mwFunc ...MwFunc)
}

type SvrOpt func(*HttpSvr)

type HttpSvr struct {
	*routeTree

	addr string
}

func NewHttpSvr(opts ...SvrOpt) *HttpSvr {
	svr := &HttpSvr{
		routeTree: newRouteTree(),
		addr:      ":8080",
	}

	for _, opt := range opts {
		opt(svr)
	}

	return svr
}

func SvrWithAddr(addr string) SvrOpt {
	return func(s *HttpSvr) {
		s.addr = addr
	}
}

func (s *HttpSvr) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := &Context{
		Req: r,
		Rsp: w,
	}

	s.serve(ctx)
}

// serve is the main function to serve the request
func (s *HttpSvr) serve(ctx *Context) {
	matched := s.getRoute(ctx.Req.Method, ctx.Req.URL.Path)
	if matched.node == nil {
		ctx.RspJson(http.StatusNotFound, "Not Found")
		return
	}

	ctx.matchedPath = matched.node.fullRoute

	hdlFunc := matched.node.hdlFunc
	// middleware execution
	mwChain := matched.node.mwChain
	// reverse the middleware chain
	for i := len(mwChain) - 1; i >= 0; i-- {
		hdlFunc = mwChain[i](hdlFunc)
	}

	// wrap the handler function
	// flush the response after the handler function is executed
	hdlFunc = func(next HdlFunc) HdlFunc {
		return func(ctx *Context) {
			next(ctx)
			s.flushRsp(ctx)
		}
	}(hdlFunc)

	ctx.pathParams = matched.params
	hdlFunc(ctx)
}

func (h *HttpSvr) flushRsp(ctx *Context) {
	if ctx.StatusCode > 0 {
		ctx.Rsp.WriteHeader(ctx.StatusCode)
	}

	if _, err := ctx.Rsp.Write(ctx.Data); err != nil {
		log.Fatalln("[easy_web] flush response failed", err)
	}
}

func (s *HttpSvr) Start() error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}

	return http.Serve(ln, s)
}

func (s *HttpSvr) RouteRegister(method string, path string, hdl HdlFunc, mwFunc ...MwFunc) {
	s.addRoute(method, path, hdl, mwFunc...)
}

func (s *HttpSvr) Get(path string, hdl HdlFunc, mwFunc ...MwFunc) {
	s.addRoute(http.MethodGet, path, hdl, mwFunc...)
}

func (s *HttpSvr) Post(path string, hdl HdlFunc, mwFunc ...MwFunc) {
	s.addRoute(http.MethodPost, path, hdl, mwFunc...)
}

func (s *HttpSvr) Put(path string, hdl HdlFunc, mwFunc ...MwFunc) {
	s.addRoute(http.MethodPut, path, hdl, mwFunc...)
}

func (s *HttpSvr) Patch(path string, hdl HdlFunc, mwFunc ...MwFunc) {
	s.addRoute(http.MethodPatch, path, hdl, mwFunc...)
}

func (s *HttpSvr) Delete(path string, hdl HdlFunc, mwFunc ...MwFunc) {
	s.addRoute(http.MethodDelete, path, hdl, mwFunc...)
}

func (s *HttpSvr) Head(path string, hdl HdlFunc, mwFunc ...MwFunc) {
	s.addRoute(http.MethodHead, path, hdl, mwFunc...)
}

func (s *HttpSvr) Options(path string, hdl HdlFunc, mwFunc ...MwFunc) {
	s.addRoute(http.MethodOptions, path, hdl, mwFunc...)
}

func (s *HttpSvr) Group(path string) *RouteGroup {
	return newRouteGroup(s, path)
}

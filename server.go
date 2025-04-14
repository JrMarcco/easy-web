package easy_web

import (
	"net"
	"net/http"
)

var _ Server = (*HttpSvr)(nil)

type HdlFunc func(ctx *Context)

type Server interface {
	http.Handler

	Start() error
	RouteRegister(method string, path string, hdl HdlFunc)
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

func (s *HttpSvr) serve(ctx *Context) {
	matched := s.getRoute(ctx.Req.Method, ctx.Req.URL.Path)
	if !matched.ok {
		ctx.RspJson(http.StatusNotFound, "Not Found")
		return
	}

	ctx.pathParams = matched.params

	hdlFunc := matched.hdlFunc
	hdlFunc(ctx)
}

func (s *HttpSvr) Start() error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}

	return http.Serve(ln, s)
}

func (s *HttpSvr) RouteRegister(method string, path string, hdl HdlFunc) {
	s.addRoute(method, path, hdl)
}

func (s *HttpSvr) Get(path string, hdl HdlFunc) {
	s.addRoute(http.MethodGet, path, hdl)
}

func (s *HttpSvr) Post(path string, hdl HdlFunc) {
	s.addRoute(http.MethodPost, path, hdl)
}

func (s *HttpSvr) Put(path string, hdl HdlFunc) {
	s.addRoute(http.MethodPut, path, hdl)
}

func (s *HttpSvr) Patch(path string, hdl HdlFunc) {
	s.addRoute(http.MethodPatch, path, hdl)
}

func (s *HttpSvr) Delete(path string, hdl HdlFunc) {
	s.addRoute(http.MethodDelete, path, hdl)
}

func (s *HttpSvr) Head(path string, hdl HdlFunc) {
	s.addRoute(http.MethodHead, path, hdl)
}

func (s *HttpSvr) Options(path string, hdl HdlFunc) {
	s.addRoute(http.MethodOptions, path, hdl)
}

func (s *HttpSvr) Group(path string) *RouteGroup {
	return newRouteGroup(s, path)
}

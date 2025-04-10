package easy_web

import "net/http"

var _ Server = (*HttpSvr)(nil)

type HdlFunc func(ctx *Context)

type Server interface {
	http.Handler

	Start(addr string) error
	RouteRegister(method string, path string, hdl HdlFunc)
}

type HttpSvr struct {
	*routeTree
}

func NewHttpSvr() *HttpSvr {
	return &HttpSvr{
		routeTree: newRouteTree(),
	}
}

func (s *HttpSvr) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := &Context{
		Req: r,
		Rsp: w,
	}

	// TODO
	s.serve(ctx)
}

func (s *HttpSvr) serve(ctx *Context) {
	panic("not implemented")
}

func (s *HttpSvr) Start(addr string) error {
	panic("not implemented")
}

func (s *HttpSvr) RouteRegister(method string, path string, hdl HdlFunc) {
	panic("not implemented")
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

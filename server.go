package easy_web

import "net/http"

var _ Server = (*HttpSvr)(nil)

type HdlFunc func(ctx *Context)

type Server interface {
	http.Handler

	Start(addr string) error
	RouteRegister(method string, path string, hdl HdlFunc)
}

type HttpSvr struct{}

func (s *HttpSvr) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := &Context{
		Req:  r,
		Resp: w,
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
	s.RouteRegister(http.MethodGet, path, hdl)
}

func (s *HttpSvr) Post(path string, hdl HdlFunc) {
	s.RouteRegister(http.MethodPost, path, hdl)
}

func (s *HttpSvr) Put(path string, hdl HdlFunc) {
	s.RouteRegister(http.MethodPut, path, hdl)
}

func (s *HttpSvr) Delete(path string, hdl HdlFunc) {
	s.RouteRegister(http.MethodDelete, path, hdl)
}

func (s *HttpSvr) Patch(path string, hdl HdlFunc) {
	s.RouteRegister(http.MethodPatch, path, hdl)
}

func (s *HttpSvr) Head(path string, hdl HdlFunc) {
	s.RouteRegister(http.MethodHead, path, hdl)
}

func (s *HttpSvr) Options(path string, hdl HdlFunc) {
	s.RouteRegister(http.MethodOptions, path, hdl)
}

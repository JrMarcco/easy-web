package easyweb

import (
	"log"
	"net"
	"net/http"
)

var _ Server = (*HttpServer)(nil)

// HandleFunc is a handler function for a route
type HandleFunc func(ctx *Context)

// Middleware is a middleware function that returns a new HdlFunc
// It is an aop implementation.
type Middleware func(next HandleFunc) HandleFunc

// MiddlewareChain is a chain of middleware functions
type MiddlewareChain []Middleware

type Server interface {
	http.Handler

	Start() error
	RouteRegister(method string, path string, hdl HandleFunc, mwFunc ...Middleware)
}

type ServerOpt func(*HttpServer)

type HttpServer struct {
	*routeTree

	addr string
}

func NewHttpServer(opts ...ServerOpt) *HttpServer {
	svr := &HttpServer{
		routeTree: newRouteTree(),
		addr:      ":8080",
	}

	for _, opt := range opts {
		opt(svr)
	}

	return svr
}

func WithAddr(addr string) ServerOpt {
	return func(s *HttpServer) {
		s.addr = addr
	}
}

func (s *HttpServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := &Context{
		Req:      r,
		Resp:     w,
		TraceCtx: r.Context(),
	}

	s.serve(ctx)
}

// serve is the main function to serve the request
func (s *HttpServer) serve(ctx *Context) {
	matched := s.getRoute(ctx.Req.Method, ctx.Req.URL.Path)
	defer s.putMatchInfo(matched)

	if matched.node == nil {
		_ = ctx.RespBytes(http.StatusNotFound, []byte("Not Found"))
		return
	}

	ctx.MatchedRoute = matched.node.fullRoute

	handleFunc := matched.node.handleFunc
	// middleware execution
	middlewareChain := matched.node.middlewareChain
	// reverse the middleware chain
	for i := len(middlewareChain) - 1; i >= 0; i-- {
		handleFunc = middlewareChain[i](handleFunc)
	}

	// wrap the handler function
	// flush the response after the handler function is executed
	handleFunc = func(next HandleFunc) HandleFunc {
		return func(ctx *Context) {
			next(ctx)
			s.flushResp(ctx)
		}
	}(handleFunc)

	ctx.pathParams = matched.params
	handleFunc(ctx)
}

func (s *HttpServer) flushResp(ctx *Context) {
	if ctx.StatusCode > 0 {
		ctx.Resp.WriteHeader(ctx.StatusCode)
	}

	if _, err := ctx.Resp.Write(ctx.Data); err != nil {
		log.Fatalln("[easy_web] flush response failed", err)
	}
}

func (s *HttpServer) Start() error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}

	return http.Serve(ln, s)
}

func (s *HttpServer) RouteRegister(method string, path string, hdl HandleFunc, mwFunc ...Middleware) {
	s.addRoute(method, path, hdl, mwFunc...)
}

func (s *HttpServer) Get(path string, hdl HandleFunc, mwFunc ...Middleware) {
	s.addRoute(http.MethodGet, path, hdl, mwFunc...)
}

func (s *HttpServer) Post(path string, hdl HandleFunc, mwFunc ...Middleware) {
	s.addRoute(http.MethodPost, path, hdl, mwFunc...)
}

func (s *HttpServer) Put(path string, hdl HandleFunc, mwFunc ...Middleware) {
	s.addRoute(http.MethodPut, path, hdl, mwFunc...)
}

func (s *HttpServer) Patch(path string, hdl HandleFunc, mwFunc ...Middleware) {
	s.addRoute(http.MethodPatch, path, hdl, mwFunc...)
}

func (s *HttpServer) Delete(path string, hdl HandleFunc, mwFunc ...Middleware) {
	s.addRoute(http.MethodDelete, path, hdl, mwFunc...)
}

func (s *HttpServer) Head(path string, hdl HandleFunc, mwFunc ...Middleware) {
	s.addRoute(http.MethodHead, path, hdl, mwFunc...)
}

func (s *HttpServer) Options(path string, hdl HandleFunc, mwFunc ...Middleware) {
	s.addRoute(http.MethodOptions, path, hdl, mwFunc...)
}

func (s *HttpServer) Group(path string) *RouteGroup {
	return newRouteGroup(s, path)
}

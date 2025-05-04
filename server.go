package easyweb

import (
	"log"
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
	Route(method string, path string, hdl HandleFunc, mwFunc ...Middleware)
}

type HttpServer struct {
	*routeTree

	addr      string
	tplEngine TemplateEngine
}

type ServerOpt func(*HttpServer)

func ServerWithAddrOpt(addr string) ServerOpt {
	return func(s *HttpServer) {
		s.addr = addr
	}
}

func ServerWithTplEngineOpt(tplEngine TemplateEngine) ServerOpt {
	return func(s *HttpServer) {
		s.tplEngine = tplEngine
	}
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

func (s *HttpServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := &Context{
		Req:       r,
		Resp:      w,
		TraceCtx:  r.Context(),
		tplEngine: s.tplEngine,
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
	return http.ListenAndServe(s.addr, s)
}

func (s *HttpServer) Route(method string, path string, hdl HandleFunc, mws ...Middleware) {
	s.addRoute(method, path, hdl, mws...)
}

func (s *HttpServer) Group(path string) *RouteGroup {
	return newRouteGroup(s, path)
}

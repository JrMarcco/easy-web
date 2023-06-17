package web

import (
	"log"
	"net"
	"net/http"
)

type HandleFunc func(*Context)

type Middleware func(next HandleFunc) HandleFunc
type MiddlewareChain []Middleware

type HttpServer struct {
	*router

	addr string
}

type SvrOpt func(server *HttpServer)

func NewHttpServer(opts ...SvrOpt) *HttpServer {

	server := &HttpServer{
		addr:   ":8080",
		router: newRouter(),
	}

	for _, opt := range opts {
		opt(server)
	}

	return server
}

func SvrWithAddr(addr string) SvrOpt {
	return func(server *HttpServer) {
		server.addr = addr
	}
}

func (h *HttpServer) Start() error {
	l, err := net.Listen("tcp", h.addr)
	if err != nil {
		return err
	}

	return http.Serve(l, h)
}

func (h *HttpServer) Get(path string, handleFunc HandleFunc, middlewares ...Middleware) {
	h.addRoute(http.MethodGet, path, handleFunc, middlewares)
}

func (h *HttpServer) Post(path string, handleFunc HandleFunc, middlewares ...Middleware) {
	h.addRoute(http.MethodPost, path, handleFunc, middlewares)
}

func (h *HttpServer) Put(path string, handleFunc HandleFunc, middlewares ...Middleware) {
	h.addRoute(http.MethodPut, path, handleFunc, middlewares)
}

func (h *HttpServer) Delete(path string, handleFunc HandleFunc, middlewares ...Middleware) {
	h.addRoute(http.MethodDelete, path, handleFunc, middlewares)
}

func (h *HttpServer) Group(prefix string) *Group {
	return newGroup(h, prefix)
}

func (h *HttpServer) ServeHTTP(rsp http.ResponseWriter, req *http.Request) {

	ctx := &Context{
		Req:      req,
		Rsp:      rsp,
		TraceCtx: req.Context(),
	}

	mi := h.findRoute(ctx.Req.Method, ctx.Req.URL.Path)
	if !mi.matched {
		ctx.RspStatusCode = http.StatusNotFound
		ctx.RspData = []byte("Not Found")
		return
	}

	handleFunc := mi.handleFunc
	for i := len(mi.middleWares) - 1; i >= 0; i-- {
		handleFunc = mi.middleWares[i](handleFunc)
	}

	handleFunc = func(next HandleFunc) HandleFunc {
		return func(context *Context) {
			next(ctx)
			h.flushRsp(ctx)
		}
	}(handleFunc)

	ctx.MatchedRoute = mi.matchedRoute
	ctx.pathParams = mi.params
	handleFunc(ctx)
}

func (h *HttpServer) flushRsp(ctx *Context) {
	if ctx.RspStatusCode > 0 {
		ctx.Rsp.WriteHeader(ctx.RspStatusCode)
	}

	_, err := ctx.Rsp.Write(ctx.RspData)
	if err != nil {
		log.Fatalln("[server] write rsp data error", err)
	}
}

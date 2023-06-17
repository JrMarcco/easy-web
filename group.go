package web

type Group struct {
	server      *HttpServer
	prefix      string
	parent      *Group
	middlewares MiddlewareChain
}

func newGroup(server *HttpServer, prefix string) *Group {

	if prefix[0] != '/' {
		panic("[route] group prefix must start with '/'")
	}

	return &Group{
		server: server,
		prefix: prefix,
	}
}

func (g *Group) Group(prefix string) *Group {
	if prefix[0] != '/' {
		panic("[route] group prefix must start with '/'")
	}

	child := newGroup(g.server, prefix)
	child.parent = g

	return child
}

func (g *Group) Get(path string, handleFunc HandleFunc, middlewares ...Middleware) {
	all := append(g.getMiddlewares(), middlewares...)
	g.server.Get(g.GetAbsolutePrefix()+path, handleFunc, all...)
}

func (g *Group) Post(path string, handleFunc HandleFunc, middlewares ...Middleware) {
	all := append(g.getMiddlewares(), middlewares...)
	g.server.Post(g.GetAbsolutePrefix()+path, handleFunc, all...)
}

func (g *Group) Put(path string, handleFunc HandleFunc, middlewares ...Middleware) {
	all := append(g.getMiddlewares(), middlewares...)
	g.server.Put(g.GetAbsolutePrefix()+path, handleFunc, all...)
}

func (g *Group) Delete(path string, handleFunc HandleFunc, middlewares ...Middleware) {
	all := append(g.getMiddlewares(), middlewares...)
	g.server.Delete(g.GetAbsolutePrefix()+path, handleFunc, all...)
}

func (g *Group) Use(middlewares ...Middleware) {
	g.middlewares = append(g.middlewares, middlewares...)
}

func (g *Group) getMiddlewares() MiddlewareChain {
	if g.parent == nil {
		return g.middlewares
	}
	return append(g.parent.middlewares, g.middlewares...)
}

func (g *Group) GetAbsolutePrefix() string {
	if g.parent == nil {
		return g.prefix
	}

	return g.parent.GetAbsolutePrefix() + g.prefix
}

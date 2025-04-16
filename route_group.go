package easyweb

type RouteGroup struct {
	svr *HttpServer

	basePath string
	parent   *RouteGroup

	mwChain MiddlewareChain
}

func newRouteGroup(svr *HttpServer, path string) *RouteGroup {
	if path == "" || path[0] != '/' {
		panic("[easy_web] path must start with '/'")
	}

	return &RouteGroup{
		svr:      svr,
		basePath: path,
	}
}

func (rg *RouteGroup) Group(prefix string) *RouteGroup {
	newRg := newRouteGroup(rg.svr, prefix)
	newRg.parent = rg

	return newRg
}

func (rg *RouteGroup) Get(relativePath string, hdlFunc HandleFunc, mws ...Middleware) {
	mwChain := append(rg.getMwChain(), mws...)
	rg.svr.Get(rg.getAbsPath()+relativePath, hdlFunc, mwChain...)
}

func (rg *RouteGroup) Post(relativePath string, hdlFunc HandleFunc, mws ...Middleware) {
	mwChain := append(rg.getMwChain(), mws...)
	rg.svr.Post(rg.getAbsPath()+relativePath, hdlFunc, mwChain...)
}

func (rg *RouteGroup) Put(relativePath string, hdlFunc HandleFunc, mws ...Middleware) {
	mwChain := append(rg.getMwChain(), mws...)
	rg.svr.Put(rg.getAbsPath()+relativePath, hdlFunc, mwChain...)
}

func (rg *RouteGroup) Patch(relativePath string, hdlFunc HandleFunc, mws ...Middleware) {
	mwChain := append(rg.getMwChain(), mws...)
	rg.svr.Patch(rg.getAbsPath()+relativePath, hdlFunc, mwChain...)
}

func (rg *RouteGroup) Delete(relativePath string, hdlFunc HandleFunc, mws ...Middleware) {
	mwChain := append(rg.getMwChain(), mws...)
	rg.svr.Delete(rg.getAbsPath()+relativePath, hdlFunc, mwChain...)
}

func (rg *RouteGroup) Head(relativePath string, hdlFunc HandleFunc, mws ...Middleware) {
	mwChain := append(rg.getMwChain(), mws...)
	rg.svr.Head(rg.getAbsPath()+relativePath, hdlFunc, mwChain...)
}

func (rg *RouteGroup) Options(relativePath string, hdlFunc HandleFunc, mws ...Middleware) {
	mwChain := append(rg.getMwChain(), mws...)
	rg.svr.Options(rg.getAbsPath()+relativePath, hdlFunc, mwChain...)
}

func (rg *RouteGroup) getAbsPath() string {
	if rg.parent == nil {
		return rg.basePath
	}

	return rg.parent.getAbsPath() + rg.basePath
}

func (rg *RouteGroup) Use(mw ...Middleware) {
	rg.mwChain = append(rg.mwChain, mw...)
}

func (rg *RouteGroup) getMwChain() MiddlewareChain {
	if rg.parent == nil {
		return rg.mwChain
	}

	return append(rg.parent.getMwChain(), rg.mwChain...)
}

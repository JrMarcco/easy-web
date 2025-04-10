package easy_web

type RouteGroup struct {
	svr *HttpSvr

	basePath string
	parent   *RouteGroup
}

func newRouteGroup(svr *HttpSvr, path string) *RouteGroup {
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

func (rg *RouteGroup) Get(relativePath string, hdlFunc HdlFunc) {
	rg.svr.Get(rg.getAbsPath()+relativePath, hdlFunc)
}

func (rg *RouteGroup) Post(relativePath string, hdlFunc HdlFunc) {
	rg.svr.Post(rg.getAbsPath()+relativePath, hdlFunc)
}

func (rg *RouteGroup) Put(relativePath string, hdlFunc HdlFunc) {
	rg.svr.Put(rg.getAbsPath()+relativePath, hdlFunc)
}

func (rg *RouteGroup) Patch(relativePath string, hdlFunc HdlFunc) {
	rg.svr.Patch(rg.getAbsPath()+relativePath, hdlFunc)
}

func (rg *RouteGroup) Delete(relativePath string, hdlFunc HdlFunc) {
	rg.svr.Delete(rg.getAbsPath()+relativePath, hdlFunc)
}

func (rg *RouteGroup) Head(relativePath string, hdlFunc HdlFunc) {
	rg.svr.Head(rg.getAbsPath()+relativePath, hdlFunc)
}

func (rg *RouteGroup) Options(relativePath string, hdlFunc HdlFunc) {
	rg.svr.Options(rg.getAbsPath()+relativePath, hdlFunc)
}

func (rg *RouteGroup) getAbsPath() string {
	if rg.parent == nil {
		return rg.basePath
	}

	return rg.parent.getAbsPath() + rg.basePath
}

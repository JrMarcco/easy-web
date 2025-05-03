package easyweb

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
)

type routeTree struct {
	m    map[string]*node
	pool sync.Pool
}

func newRouteTree() *routeTree {
	return &routeTree{
		m: make(map[string]*node),
		pool: sync.Pool{
			New: func() any {
				return &matched{
					params: make(map[string]string),
				}
			},
		},
	}
}

func (t *routeTree) addRoute(method string, path string, hdlFunc HandleFunc, mws ...Middleware) {
	if path == "" {
		panic("[easy_web] path is empty")
	}

	if path[0] != '/' {
		panic("[easy_web] path must start with '/'")
	}

	root, ok := t.m[method]
	if !ok {
		root = &node{}
		t.m[method] = root
	}

	if path == "/" {
		if root.handleFunc != nil {
			panic(fmt.Sprintf("[easy_web] route %s already exists", path))
		}

		root.fullRoute = path
		root.handleFunc = hdlFunc
		root.middlewareChain = append(root.middlewareChain, mws...)
		return
	}

	segments := strings.SplitSeq(strings.Trim(path, "/"), "/")
	for seg := range segments {
		if seg == "" {
			panic("[easy_web] path contains consecutive '/'")
		}

		root = root.addChild(seg)
	}

	if root.handleFunc != nil {
		panic(fmt.Sprintf("[easy_web] route %s already exists", path))
	}

	root.fullRoute = strings.TrimRight(path, "/")
	root.handleFunc = hdlFunc
	root.middlewareChain = append(root.middlewareChain, mws...)
}

func (t *routeTree) getRoute(method string, path string) *matched {
	matched := t.pool.Get().(*matched)

	root, ok := t.m[method]
	if !ok {
		return matched
	}

	path = strings.Trim(path, "/")
	if path == "" {
		matched.node = root
		return matched
	}

	segments := strings.SplitSeq(path, "/")
	isWildcardParent := false
	for seg := range segments {
		child, ok := root.getChild(seg)
		if !ok {
			if isWildcardParent {
				continue
			}
			return matched
		}

		// cache path params
		if child.typ == param {
			matched.addParam(child.baseRoute[1:], seg)
		}

		if child.typ == wildcard {
			isWildcardParent = true
		} else {
			isWildcardParent = false
		}
		root = child
	}

	matched.node = root
	return matched
}

// putMatchInfo returns a matchInfo to the pool
func (t *routeTree) putMatchInfo(matched *matched) {
	// Reset the matchInfo before put back to pool
	matched.reset()

	t.pool.Put(matched)
}

const (
	static   = iota // static route node ( e.g. /mall/order )
	wildcard        // wildcard route node ( e.g. /mall/* )
	param           // param route node ( e.g., /mall/order/:id )
	reg             // regular exp node ( e.g., /mall/order/re:^\d+$ )
)

type nodeType int8

type node struct {
	typ       nodeType
	baseRoute string
	fullRoute string
	children  map[string]*node
	wildcardN *node
	paramN    *node
	regexpN   *node

	re              *regexp.Regexp
	handleFunc      HandleFunc
	middlewareChain MiddlewareChain
}

func (n *node) addChild(path string) *node {
	if path == "*" {
		return n.addWildcardN()
	}

	if path[0] == ':' {
		return n.addParamN(path)
	}

	if strings.HasPrefix(path, "re:") {
		return n.addRegexpN(path)
	}

	if n.children == nil {
		n.children = make(map[string]*node)
	}

	// if exists, return
	if child, ok := n.children[path]; ok {
		return child
	}

	// create a new node
	newNode := &node{
		baseRoute: path,
		typ:       static,
	}
	n.children[path] = newNode
	return newNode
}

func (n *node) addWildcardN() *node {
	if n.wildcardN == nil {
		if n.paramN != nil || n.regexpN != nil {
			panic("[easy_web] can not register wildcard/param/regexp node at the same time")
		}

		n.wildcardN = &node{
			typ:       wildcard,
			baseRoute: "*",
		}
	}

	return n.wildcardN
}

func (n *node) addParamN(path string) *node {
	if n.wildcardN != nil || n.regexpN != nil {
		panic("[easy_web] can not register wildcard/param/regexp node at the same time")
	}

	if n.paramN != nil {
		if n.paramN.baseRoute != path {
			panic(fmt.Sprintf("[easy_web] duplicate registered param node at %s", path))
		}
		return n.paramN
	}

	n.paramN = &node{
		typ:       param,
		baseRoute: path,
	}
	return n.paramN
}

func (n *node) addRegexpN(path string) *node {
	if n.wildcardN != nil || n.paramN != nil {
		panic("[easy_web] can not register wildcard/param/regexp node at the same time")
	}

	re := regexp.MustCompile(path[3:])
	if n.regexpN != nil {
		if n.regexpN.re.String() != re.String() {
			panic(fmt.Sprintf("[easy_web] duplicate registered regexp node at %s", path))
		}
		return n.regexpN
	}

	n.regexpN = &node{
		typ:       reg,
		baseRoute: path,
		re:        re,
	}
	return n.regexpN
}

func (n *node) getChild(path string) (*node, bool) {
	// check the regexp node first
	if n.regexpN != nil && n.regexpN.re.MatchString(path) {
		return n.regexpN, true
	}

	// check wildcard or param node
	child, ok := n.children[path]
	if ok {
		return child, true
	}
	return n.getWildcardOrParamN()
}

func (n *node) getWildcardOrParamN() (*node, bool) {
	if n.wildcardN != nil {
		return n.wildcardN, true
	}

	if n.paramN != nil {
		return n.paramN, true
	}

	return nil, false
}

type matched struct {
	node   *node
	params map[string]string
}

func (m *matched) addParam(key, value string) {
	m.params[key] = value
}

func (m *matched) reset() {
	m.node = nil
	for k := range m.params {
		delete(m.params, k)
	}
}

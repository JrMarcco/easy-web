package web

import (
	"fmt"
	"strings"
)

type router struct {
	methodTrees map[string]*node
}

func newRouter() *router {
	return &router{
		methodTrees: map[string]*node{},
	}
}

func (r *router) addRoute(method string, path string, handleFunc HandleFunc, middlewares MiddlewareChain) {

	if path == "" {
		panic("[route] empty path")
	}

	if path[0] != '/' {
		panic("[route] path must start with '/'")
	}

	root, ok := r.methodTrees[method]
	if !ok {
		root = &node{path: "/"}
		r.methodTrees[method] = root
	}

	if path == "/" {
		if root.handleFunc != nil {
			panic(fmt.Sprintf("[route] path '%s' has already registered", path))
		}
		root.handleFunc = handleFunc
		root.middlewares = middlewares
		return
	}

	segs := strings.Split(strings.Trim(path, "/"), "/")
	for _, seg := range segs {
		root = root.createChild(seg)
	}

	if root.handleFunc != nil {
		panic(fmt.Sprintf("[route] path '%s' has already registered", path))
	}
	root.handleFunc = handleFunc
	root.middlewares = middlewares
	root.completeRoute = path
}

func (r *router) findRoute(method string, path string) matchInfo {

	root, ok := r.methodTrees[method]
	if !ok {
		return matchInfo{matched: false}
	}

	if path == "/" {
		if root.handleFunc == nil {
			return matchInfo{matched: false}
		}
		return matchInfo{
			matched:      true,
			matchedRoute: "/",
			handleFunc:   root.handleFunc,
		}
	}

	var params map[string]string

	segs := strings.Split(strings.Trim(path, "/"), "/")
	for _, seg := range segs {
		child, ok := root.findChild(seg)
		if !ok {
			if root.typ == nodeTypeWildcard {
				return matchInfo{
					matched:    true,
					handleFunc: root.handleFunc,
					params:     params,
				}
			}
			return matchInfo{matched: false}
		}

		// 获取参数路径上的参数
		if child.path[0] == ':' {
			if params == nil {
				params = make(map[string]string, 2)
			}

			params[child.path[1:]] = seg
		}

		root = child
	}

	return matchInfo{
		matched:      root.handleFunc != nil,
		matchedRoute: root.completeRoute,
		handleFunc:   root.handleFunc,
		middleWares:  root.middlewares,
		params:       params,
	}
}

const (
	nodeTypeStatic = iota
	nodeTypeWildcard
	nodeTypeParam
)

type nodeType int

type node struct {
	typ           nodeType
	path          string
	completeRoute string
	children      map[string]*node
	wildcardNode  *node
	paramNode     *node
	handleFunc    HandleFunc
	middlewares   MiddlewareChain
}

func (n *node) createChild(path string) *node {
	if path == "*" {
		return n.createWildCardNode(path)
	}

	if path[0] == ':' {
		return n.createParamNode(path)
	}

	if n.children == nil {
		n.children = map[string]*node{}
	}

	if child, ok := n.children[path]; ok {
		return child
	}

	n.children[path] = &node{
		typ:  nodeTypeStatic,
		path: path,
	}
	return n.children[path]
}

func (n *node) createWildCardNode(path string) *node {
	if n.wildcardNode == nil {
		if n.paramNode != nil {
			panic("[route] can not register wildcardNode and paramNode at same time")
		}

		n.wildcardNode = &node{
			typ:  nodeTypeWildcard,
			path: path,
		}
	}

	return n.wildcardNode
}

func (n *node) createParamNode(path string) *node {
	if n.wildcardNode != nil {
		panic("[route] can not register wildcardNode and paramNode at same time")
	}
	n.paramNode = &node{
		typ:  nodeTypeParam,
		path: path,
	}
	return n.paramNode
}

func (n *node) findChild(path string) (*node, bool) {

	if n.children == nil {
		return n.findSpecNode()
	}

	if child, ok := n.children[path]; ok {
		return child, true
	}

	return n.findSpecNode()
}

func (n *node) findSpecNode() (*node, bool) {
	if n.paramNode != nil {
		return n.paramNode, true
	}
	return n.wildcardNode, n.wildcardNode != nil
}

type matchInfo struct {
	matched      bool
	matchedRoute string
	handleFunc   HandleFunc
	middleWares  MiddlewareChain
	params       map[string]string
}

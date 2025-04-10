package easy_web

import (
	"fmt"
	"strings"
)

type routeTree struct {
	m map[string]*node
}

func newRouteTree() *routeTree {
	return &routeTree{
		m: make(map[string]*node),
	}
}

func (t *routeTree) addRoute(method string, path string, hdlFunc HdlFunc) {
	if path == "" {
		panic("[easy_web] path is empty")
	}

	if path[0] != '/' {
		panic("[easy_web] path must start with '/'")
	}

	root, ok := t.m[method]
	if !ok {
		root = &node{path: "/"}
		t.m[method] = root
	}

	path = strings.Trim(path, "/")
	if path == "" {
		if root.hdlFunc != nil {
			panic(fmt.Sprintf("[easy_web] route %s already exists", path))
		}

		root.hdlFunc = hdlFunc
		return
	}

	segments := strings.SplitSeq(strings.Trim(path, "/"), "/")
	for seg := range segments {
		if seg == "" {
			panic("[easy_web] path contains consecutive '/'")
		}

		root = root.addChild(seg)
	}

	if root.hdlFunc != nil {
		panic(fmt.Sprintf("[easy_web] route %s already exists", path))
	}

	root.hdlFunc = hdlFunc
}

func (t *routeTree) getRoute(method string, path string) (*node, bool) {
	root, ok := t.m[method]
	if !ok {
		return nil, false
	}

	path = strings.Trim(path, "/")
	if path == "" {
		return root, root.hdlFunc != nil
	}

	segments := strings.SplitSeq(strings.Trim(path, "/"), "/")
	for seg := range segments {
		child, ok := root.getChild(seg)
		if !ok {
			return nil, false
		}

		root = child
	}

	return root, root.hdlFunc != nil
}

const (
	nodeTypeStatic   = iota // static route node, e.g. /mall/order
	nodeTypeWildcard        // wildcard route node, e.g. /mall/*
	nodeTypeParam           // param route node, e.g. /mall/order/:id
)

type node struct {
	typ          int8
	path         string
	children     map[string]*node
	wildcardNode *node
	paramNode    *node

	hdlFunc HdlFunc
}

func (n *node) addChild(path string) *node {
	if path == "*" {
		return n.addWildcardNode()
	}

	if path[0] == ':' {
		return n.addParamNode(path)
	}

	if n.children == nil {
		n.children = make(map[string]*node)
	}

	// if exists, return
	if child, ok := n.children[path]; ok {
		return child
	}

	// create new node
	newNode := &node{
		path: path,
		typ:  nodeTypeStatic,
	}
	n.children[path] = newNode

	return newNode
}

func (n *node) addWildcardNode() *node {
	if n.paramNode != nil {
		panic("[easy_web] can not register wildcard node and param node at the same time")
	}

	n.wildcardNode = &node{
		path: "*",
		typ:  nodeTypeWildcard,
	}

	return n.wildcardNode
}

func (n *node) addParamNode(path string) *node {
	if n.wildcardNode != nil {
		panic("[easy_web] can not register wildcard node and param node at the same time")
	}

	n.paramNode = &node{
		path: path,
		typ:  nodeTypeParam,
	}

	return n.paramNode
}

func (n *node) getChild(path string) (*node, bool) {
	if n.children == nil {
		return nil, false
	}

	if child, ok := n.children[path]; ok {
		return child, true
	}

	return nil, false
}

package easy_web

import (
	"fmt"
	"regexp"
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

	segments := strings.SplitSeq(path, "/")
	for seg := range segments {
		if seg == "" {
			panic("[easy_web] path contains consecutive '/'")
		}

		if strings.HasPrefix(seg, "re:") {
			// regular expression node
			reNode := root.addRegChild(seg, hdlFunc)
			root.reNode = reNode
			return
		}

		root = root.addChild(seg)
	}

	if root.hdlFunc != nil {
		panic(fmt.Sprintf("[easy_web] route %s already exists", path))
	}

	root.hdlFunc = hdlFunc
}

type matchInfo struct {
	matched bool
	hdlFunc HdlFunc
	params  map[string]string
}

func (t *routeTree) getRoute(method string, path string) matchInfo {
	root, ok := t.m[method]
	if !ok {
		return matchInfo{matched: false}
	}

	path = strings.Trim(path, "/")
	if path == "" {
		return matchInfo{
			matched: root.hdlFunc != nil,
			hdlFunc: root.hdlFunc,
		}
	}

	var params map[string]string

	segments := strings.SplitSeq(path, "/")
	for seg := range segments {
		// check regular expression node first
		if root.reNode != nil {
			reNode := root.reNode
			if reNode.hdlFunc == nil || reNode.re == nil {
				return matchInfo{matched: false}
			}

			matched := reNode.re.MatchString(seg)
			return matchInfo{
				matched: matched,
				hdlFunc: reNode.hdlFunc,
			}
		}

		child, ok := root.getChild(seg)
		if !ok {
			return matchInfo{matched: false}
		}

		// cache path params
		if child.typ == param {
			if params == nil {
				params = make(map[string]string)
			}

			params[child.path[1:]] = seg
		}

		root = child
	}

	return matchInfo{
		matched: root.hdlFunc != nil,
		hdlFunc: root.hdlFunc,
		params:  params,
	}
}

const (
	static   = iota // static route node, e.g. /mall/order
	special         // have wildcard or param node, e.g. /mall/* or /mall/:id
	wildcard        // wildcard route node, e.g. /mall/*
	param           // param route node, must be last one node in path, e.g. /mall/order/:id
)

type nodeType int8

type node struct {
	typ          nodeType
	path         string
	children     map[string]*node
	wildcardNode *node
	paramNode    *node

	reNode *reNode

	hdlFunc HdlFunc
}

type reNode struct {
	re      *regexp.Regexp
	hdlFunc HdlFunc
}

func (n *node) addRegChild(pattern string, hdlFunc HdlFunc) *reNode {
	re := regexp.MustCompile(pattern[3:])
	return &reNode{
		re:      re,
		hdlFunc: hdlFunc,
	}
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
		typ:  static,
	}
	n.children[path] = newNode
	return newNode
}

func (n *node) addWildcardNode() *node {
	if n.wildcardNode != nil {
		return n.wildcardNode
	}

	if n.paramNode != nil {
		panic("[easy_web] can not register wildcard node and param node at the same time")
	}

	n.typ = special
	n.wildcardNode = &node{typ: wildcard}
	return n.wildcardNode
}

func (n *node) addParamNode(path string) *node {
	if n.paramNode != nil {
		return n.paramNode
	}

	if n.wildcardNode != nil {
		panic("[easy_web] can not register wildcard node and param node at the same time")
	}

	n.typ = special
	n.paramNode = &node{
		path: path,
		typ:  param,
	}
	return n.paramNode
}

func (n *node) getChild(path string) (*node, bool) {
	if n.children == nil {
		return n.getSpecialChild()
	}

	if child, ok := n.children[path]; ok {
		return child, true
	}

	return nil, false
}

func (n *node) getSpecialChild() (*node, bool) {
	if n.typ != special {
		return nil, false
	}

	if n.wildcardNode != nil {
		return n.wildcardNode, true
	}

	return n.paramNode, true
}

package easy_web

import (
	"fmt"
	"strings"
)

type routeTrees struct {
	m map[string]*node
}

func newRouteTrees() *routeTrees {
	return &routeTrees{
		m: make(map[string]*node),
	}
}

func (t *routeTrees) addRoute(method string, path string, hdlFunc HdlFunc) {

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

	segments := strings.Split(strings.Trim(path, "/"), "/")

	for _, seg := range segments {
		root = root.addChild(seg)
	}

	if root.hdlFunc != nil {
		panic(fmt.Sprintf("[easy_web] route %s already exists", path))
	}

	root.hdlFunc = hdlFunc
}

const (
	nodeTypeStatic = iota
)

type node struct {
	typ      int
	path     string
	children map[string]*node
	hdlFunc  HdlFunc
}

func (n *node) addChild(path string) *node {
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

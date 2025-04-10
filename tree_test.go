package easy_web

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRouteTree_AddRoute(t *testing.T) {
	mockHdlFunc := func(ctx *Context) {}

	tcs := []struct {
		name      string
		method    string
		path      string
		wantTrees *routeTree
	}{
		{
			name:   "root node",
			method: http.MethodGet,
			path:   "/",
			wantTrees: &routeTree{
				m: map[string]*node{
					http.MethodGet: {
						typ:      nodeTypeStatic,
						path:     "/",
						hdlFunc:  mockHdlFunc,
						children: nil,
					},
				},
			},
		}, {
			name:   "basic",
			method: http.MethodGet,
			path:   "/user/test",
			wantTrees: &routeTree{
				m: map[string]*node{
					http.MethodGet: {
						typ:  nodeTypeStatic,
						path: "/",
						children: map[string]*node{
							"user": {
								typ:  nodeTypeStatic,
								path: "user",
								children: map[string]*node{
									"test": {
										typ:      nodeTypeStatic,
										path:     "test",
										hdlFunc:  mockHdlFunc,
										children: nil,
									},
								},
							},
						},
					},
				},
			},
		}, {
			name:   "path end with sprit",
			method: http.MethodGet,
			path:   "/user/test/",
			wantTrees: &routeTree{
				m: map[string]*node{
					http.MethodGet: {
						typ:  nodeTypeStatic,
						path: "/",
						children: map[string]*node{
							"user": {
								typ:  nodeTypeStatic,
								path: "user",
								children: map[string]*node{
									"test": {
										typ:      nodeTypeStatic,
										path:     "test",
										hdlFunc:  mockHdlFunc,
										children: nil,
									},
								},
							},
						},
					},
				},
			},
		}, {
			name:   "single path",
			method: http.MethodGet,
			path:   "/user",
			wantTrees: &routeTree{
				m: map[string]*node{
					http.MethodGet: {
						typ:  nodeTypeStatic,
						path: "/",
						children: map[string]*node{
							"user": {
								typ:      nodeTypeStatic,
								path:     "user",
								hdlFunc:  mockHdlFunc,
								children: nil,
							},
						},
					},
				},
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			tree := newRouteTree()
			tree.addRoute(tc.method, tc.path, mockHdlFunc)

			msg, ok := tree.equal(tc.wantTrees)
			if !ok {
				t.Log(msg)
			}
			assert.True(t, ok)
		})
	}

	// invalid path
	tree := newRouteTree()
	assert.Panics(t, func() {
		tree.addRoute(http.MethodGet, "", mockHdlFunc)
	})
	assert.Panics(t, func() {
		tree.addRoute(http.MethodGet, "user", mockHdlFunc)
	})
	assert.Panics(t, func() {
		tree.addRoute(http.MethodGet, "user//test", mockHdlFunc)
	})

	// duplicate register root path
	tree.addRoute(http.MethodGet, "/", mockHdlFunc)
	assert.Panics(t, func() {
		tree.addRoute(http.MethodGet, "/", mockHdlFunc)
	})

	// duplicate register path
	tree.addRoute(http.MethodGet, "/user/test", mockHdlFunc)
	assert.Panics(t, func() {
		tree.addRoute(http.MethodGet, "/user/test", mockHdlFunc)
	})
}

func TestRouteTree_GetRoute(t *testing.T) {
	mockHdlFunc := func(ctx *Context) {}

	tree := newRouteTree()
	tree.addRoute(http.MethodGet, "/", mockHdlFunc)

	tree.addRoute(http.MethodGet, "/v1/user", mockHdlFunc)

	tree.addRoute(http.MethodGet, "/v2/mall/order", mockHdlFunc)
	tree.addRoute(http.MethodPost, "/v2/mall/order", mockHdlFunc)

	tcs := []struct {
		name     string
		method   string
		path     string
		wantRes  bool
		wantNode *node
	}{
		{
			name:    "root node",
			method:  http.MethodGet,
			path:    "/",
			wantRes: true,
			wantNode: &node{
				typ:     nodeTypeStatic,
				path:    "/",
				hdlFunc: mockHdlFunc,
			},
		}, {
			name:     "not found",
			method:   http.MethodGet,
			path:     "/user",
			wantRes:  false,
			wantNode: nil,
		}, {
			name:    "node without hdlFunc",
			method:  http.MethodGet,
			path:    "/v2/mall",
			wantRes: false,
		}, {
			name:    "normal",
			method:  http.MethodPost,
			path:    "/v2/mall/order",
			wantRes: true,
			wantNode: &node{
				typ:     nodeTypeStatic,
				path:    "order",
				hdlFunc: mockHdlFunc,
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			node, ok := tree.getRoute(tc.method, tc.path)
			assert.Equal(t, tc.wantRes, ok)

			if ok {
				assert.Equal(t, tc.wantNode.typ, node.typ)
				assert.Equal(t, tc.wantNode.path, node.path)
				assert.True(t, tc.wantNode.hdlFunc.equal(node.hdlFunc))
			}
		})
	}
}

func (t *routeTree) equal(other *routeTree) (string, bool) {
	for method, tree := range t.m {
		otherTree, ok := other.m[method]
		if !ok {
			return fmt.Sprintf("tree not found: %s", method), false
		}

		msg, ok := tree.equal(otherTree)
		if !ok {
			return msg, false
		}
	}

	return "", true
}

func (n *node) equal(other *node) (string, bool) {
	if n.path != other.path {
		return fmt.Sprintf("path: %s != %s", n.path, other.path), false
	}

	if n.typ != other.typ {
		return fmt.Sprintf("typ: %d != %d", n.typ, other.typ), false
	}

	if len(n.children) != len(other.children) {
		return fmt.Sprintf("children: %d != %d", len(n.children), len(other.children)), false
	}

	if n.hdlFunc != nil {
		if other.hdlFunc == nil {
			return "hdlFunc not found in other", false
		}

		if !n.hdlFunc.equal(other.hdlFunc) {
			return "hdlFunc not equal", false
		}
	}

	for path, node := range n.children {
		otherNode, ok := other.children[path]
		if !ok {
			return fmt.Sprintf("other node not found: %s", path), false
		}

		msg, ok := node.equal(otherNode)
		if !ok {
			return msg, false
		}
	}

	return "", true
}

func (h HdlFunc) equal(other HdlFunc) bool {
	return reflect.ValueOf(h) == reflect.ValueOf(other)
}

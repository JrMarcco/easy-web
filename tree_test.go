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

	if len(n.children) != len(other.children) {
		return fmt.Sprintf("children: %d != %d", len(n.children), len(other.children)), false
	}

	if n.hdlFunc != nil {
		if other.hdlFunc == nil {
			return "hdlFunc not found in other", false
		}

		hdlFunc := reflect.ValueOf(n.hdlFunc)
		anotherHdlFunc := reflect.ValueOf(other.hdlFunc)

		if hdlFunc != anotherHdlFunc {
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

package web

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"reflect"
	"testing"
)

func TestRouter_AddRoute(t *testing.T) {

	tcs := []struct {
		method string
		path   string
	}{
		{
			method: http.MethodGet,
			path:   "/",
		}, {
			method: http.MethodGet,
			path:   "/home",
		}, {
			method: http.MethodGet,
			path:   "/user",
		}, {
			method: http.MethodGet,
			path:   "/user/get",
		}, {
			method: http.MethodGet,
			path:   "/user/list",
		}, {
			method: http.MethodPost,
			path:   "/user/edit",
		}, {
			method: http.MethodGet,
			path:   "/*",
		}, {
			method: http.MethodGet,
			path:   "/*/*",
		}, {
			method: http.MethodGet,
			path:   "/*/wild",
		}, {
			method: http.MethodGet,
			path:   "/*/wild/*",
		}, {
			method: http.MethodGet,
			path:   "/order/:id",
		},
	}

	r := newRouter()
	mockHandleFunc := func(ctx *Context) {}

	for _, tc := range tcs {
		r.addRoute(tc.method, tc.path, mockHandleFunc, []Middleware{})
	}

	wantRouter := &router{
		methodTrees: map[string]*node{
			http.MethodGet: {
				path:       "/",
				handleFunc: mockHandleFunc,
				children: map[string]*node{
					"home": {
						path:       "home",
						handleFunc: mockHandleFunc,
					},
					"user": {
						path:       "user",
						handleFunc: mockHandleFunc,
						children: map[string]*node{
							"get": {
								path:       "get",
								handleFunc: mockHandleFunc,
							},
							"list": {
								path:       "list",
								handleFunc: mockHandleFunc,
							},
						},
					},
					"order": {
						path: "order",
						paramNode: &node{
							path:       ":id",
							handleFunc: mockHandleFunc,
						},
					},
				},
				wildcardNode: &node{
					path:       "*",
					handleFunc: mockHandleFunc,
					children: map[string]*node{
						"wild": {
							path:       "wild",
							handleFunc: mockHandleFunc,
							wildcardNode: &node{
								path:       "*",
								handleFunc: mockHandleFunc,
							},
						},
					},
					wildcardNode: &node{
						path:       "*",
						handleFunc: mockHandleFunc,
					},
				},
			},
			http.MethodPost: {
				path: "/",
				children: map[string]*node{
					"user": {
						path: "user",
						children: map[string]*node{
							"edit": {
								path:       "edit",
								handleFunc: mockHandleFunc,
							},
						},
					},
				},
			},
		},
	}

	res, msg := wantRouter.equal(r)
	assert.True(t, res, msg)

}

func (r *router) equal(target *router) (bool, string) {
	for method, tree := range r.methodTrees {
		dst, ok := target.methodTrees[method]
		if !ok {
			return false, fmt.Sprintf("method tree '%s' unmatched", method)
		}

		res, msg := tree.equal(dst)
		if !res {
			return res, msg
		}
	}

	return true, ""
}

func (n *node) equal(target *node) (bool, string) {
	if n.path != target.path {
		return false, fmt.Sprintf("child node path '%s' unmatched", n.path)
	}

	if len(n.children) != len(target.children) {
		return false, "children length unmatched"
	}

	if n.wildcardNode != nil {
		return n.wildcardNode.equal(target.wildcardNode)
	}

	if reflect.ValueOf(n.handleFunc) != reflect.ValueOf(target.handleFunc) {
		return false, "node handle func unmatched"
	}

	for path, nd := range n.children {
		dst, ok := target.children[path]
		if !ok {
			return false, fmt.Sprintf("")
		}

		res, msg := nd.equal(dst)
		if !res {
			return res, msg
		}
	}

	return true, ""
}

func TestRouter_AddRoutePanic(t *testing.T) {
	r := newRouter()
	mockHandleFunc := func(ctx *Context) {}

	assert.PanicsWithValue(t, "[route] empty path", func() {
		r.addRoute(http.MethodGet, "", mockHandleFunc, []Middleware{})
	})

	assert.PanicsWithValue(t, "[route] path must start with '/'", func() {
		r.addRoute(http.MethodGet, "home", mockHandleFunc, []Middleware{})
	})

	registered := "/"
	r.addRoute(http.MethodGet, registered, mockHandleFunc, []Middleware{})
	assert.PanicsWithValue(
		t,
		fmt.Sprintf(fmt.Sprintf("[route] path '%s' has already registered", registered)),
		func() {
			r.addRoute(http.MethodGet, registered, mockHandleFunc, []Middleware{})
		},
	)

	registered = "/user/edit"
	r.addRoute(http.MethodPost, registered, mockHandleFunc, []Middleware{})
	assert.PanicsWithValue(
		t,
		fmt.Sprintf(fmt.Sprintf("[route] path '%s' has already registered", registered)),
		func() {
			r.addRoute(http.MethodPost, registered, mockHandleFunc, []Middleware{})
		},
	)

	r.addRoute(http.MethodGet, "/wildcardNode/*", mockHandleFunc, []Middleware{})
	assert.PanicsWithValue(
		t,
		"[route] can not register wildcardNode and paramNode at same time",
		func() {
			r.addRoute(http.MethodGet, "/wildcardNode/:id", mockHandleFunc, []Middleware{})
		},
	)

	r.addRoute(http.MethodGet, "/paramNode/:id", mockHandleFunc, []Middleware{})
	assert.PanicsWithValue(
		t,
		"[route] can not register wildcardNode and paramNode at same time",
		func() {
			r.addRoute(http.MethodGet, "/paramNode/*", mockHandleFunc, []Middleware{})
		},
	)
}

func TestRouter_findRoute(t *testing.T) {

	r := newRouter()
	mockHandleFunc := func(ctx *Context) {}

	r.addRoute(http.MethodGet, "/", mockHandleFunc, []Middleware{})
	r.addRoute(http.MethodGet, "/user/get", mockHandleFunc, []Middleware{})
	r.addRoute(http.MethodGet, "/user/list", mockHandleFunc, []Middleware{})
	r.addRoute(http.MethodPost, "/user/edit", mockHandleFunc, []Middleware{})

	r.addRoute(http.MethodGet, "/*", mockHandleFunc, []Middleware{})
	r.addRoute(http.MethodGet, "/*/*", mockHandleFunc, []Middleware{})
	r.addRoute(http.MethodGet, "/*/wild", mockHandleFunc, []Middleware{})
	r.addRoute(http.MethodGet, "/pic/*", mockHandleFunc, []Middleware{})
	r.addRoute(http.MethodGet, "/*/inner/*", mockHandleFunc, []Middleware{})

	r.addRoute(http.MethodGet, "/order/:id", mockHandleFunc, []Middleware{})
	r.addRoute(http.MethodGet, "/multi/:a/:b", mockHandleFunc, []Middleware{})

	tcs := []struct {
		name    string
		method  string
		path    string
		wantRes matchInfo
	}{
		{
			name:   "index",
			method: http.MethodGet,
			path:   "/",
			wantRes: matchInfo{
				matched:    true,
				handleFunc: mockHandleFunc,
			},
		}, {
			name:   "userGet",
			method: http.MethodGet,
			path:   "/user/get",
			wantRes: matchInfo{
				matched:    true,
				handleFunc: mockHandleFunc,
			},
		}, {
			name:   "userList",
			method: http.MethodGet,
			path:   "/user/get",
			wantRes: matchInfo{
				matched:    true,
				handleFunc: mockHandleFunc,
			},
		}, {
			name:   "userEdit",
			method: http.MethodPost,
			path:   "/user/edit",
			wantRes: matchInfo{
				matched:    true,
				handleFunc: mockHandleFunc,
			},
		}, {
			name:   "userNotFound",
			method: http.MethodPost,
			path:   "/user",
			wantRes: matchInfo{
				matched: false,
			},
		}, {
			name:   "indexWildcard",
			method: http.MethodGet,
			path:   "/demo",
			wantRes: matchInfo{
				matched:    true,
				handleFunc: mockHandleFunc,
			},
		}, {
			name:   "doubleWildcard",
			method: http.MethodGet,
			path:   "/a/b",
			wantRes: matchInfo{
				matched:    true,
				handleFunc: mockHandleFunc,
			},
		}, {
			name:   "wildcardPrefix",
			method: http.MethodGet,
			path:   "/demo/wild",
			wantRes: matchInfo{
				matched:    true,
				handleFunc: mockHandleFunc,
			},
		}, {
			name:   "wildcardNotFound",
			method: http.MethodGet,
			path:   "/a/wild/b",
			wantRes: matchInfo{
				matched: false,
			},
		}, {
			name:   "wildcardSuffix",
			method: http.MethodGet,
			path:   "/pic/abc",
			wantRes: matchInfo{
				matched:    true,
				handleFunc: mockHandleFunc,
			},
		}, {
			name:   "headAndTailWildcard1",
			method: http.MethodGet,
			path:   "/a/inner/b",
			wantRes: matchInfo{
				matched:    true,
				handleFunc: mockHandleFunc,
			},
		}, {
			name:   "headAndTailWildcard2",
			method: http.MethodGet,
			path:   "/a/inner/b/c/d",
			wantRes: matchInfo{
				matched:    true,
				handleFunc: mockHandleFunc,
			},
		}, {
			name:   "orderParam",
			method: http.MethodGet,
			path:   "/order/Z123456",
			wantRes: matchInfo{
				matched:    true,
				handleFunc: mockHandleFunc,
				params: map[string]string{
					"id": "Z123456",
				},
			},
		}, {
			name:   "multiParam",
			method: http.MethodGet,
			path:   "/multi/xxx/123",
			wantRes: matchInfo{
				matched:    true,
				handleFunc: mockHandleFunc,
				params: map[string]string{
					"a": "xxx",
					"b": "123",
				},
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			mi := r.findRoute(tc.method, tc.path)

			assert.Equal(t, tc.wantRes.matched, mi.matched)
			assert.Equal(t, tc.wantRes.params, mi.params)

			if !mi.matched {
				assert.Nil(t, mi.handleFunc)
				return
			}
			assert.True(t, reflect.ValueOf(mi.handleFunc) == reflect.ValueOf(tc.wantRes.handleFunc))
		})
	}
}

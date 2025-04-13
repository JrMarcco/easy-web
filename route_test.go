package easy_web

import (
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRouteTree_addRoute(t *testing.T) {
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
						typ:      static,
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
						typ:  static,
						path: "/",
						children: map[string]*node{
							"user": {
								typ:  static,
								path: "user",
								children: map[string]*node{
									"test": {
										typ:      static,
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
						typ:  static,
						path: "/",
						children: map[string]*node{
							"user": {
								typ:  static,
								path: "user",
								children: map[string]*node{
									"test": {
										typ:      static,
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
						typ:  static,
						path: "/",
						children: map[string]*node{
							"user": {
								typ:      static,
								path:     "user",
								hdlFunc:  mockHdlFunc,
								children: nil,
							},
						},
					},
				},
			},
		}, {
			name:   "wildcard node",
			method: http.MethodGet,
			path:   "/mall/*",
			wantTrees: &routeTree{
				m: map[string]*node{
					http.MethodGet: {
						typ:  static,
						path: "/",
						children: map[string]*node{
							"mall": {
								typ:      static,
								path:     "mall",
								children: map[string]*node{},
								wildcardN: &node{
									typ:      wildcard,
									hdlFunc:  mockHdlFunc,
									children: nil,
								},
							},
						},
					},
				},
			},
		}, {
			name:   "wildcard in the middle",
			method: http.MethodGet,
			path:   "/mall/*/transfer",
			wantTrees: &routeTree{
				m: map[string]*node{
					http.MethodGet: {
						typ:  static,
						path: "/",
						children: map[string]*node{
							"mall": {
								typ:      static,
								path:     "mall",
								children: map[string]*node{},
								wildcardN: &node{
									typ:     wildcard,
									hdlFunc: mockHdlFunc,
									children: map[string]*node{
										"transfer": {
											typ:      static,
											path:     "transfer",
											hdlFunc:  mockHdlFunc,
											children: nil,
										},
									},
								},
							},
						},
					},
				},
			},
		}, {
			name:   "param node",
			method: http.MethodGet,
			path:   "/mall/order/:id",
			wantTrees: &routeTree{
				m: map[string]*node{
					http.MethodGet: {
						typ:  static,
						path: "/",
						children: map[string]*node{
							"mall": {
								typ:  static,
								path: "mall",
								children: map[string]*node{
									"order": {
										typ:  static,
										path: "order",
										paramN: &node{
											typ:      param,
											path:     ":id",
											hdlFunc:  mockHdlFunc,
											children: nil,
										},
									},
								},
							},
						},
					},
				},
			},
		}, {
			name:   "param node in the middle",
			method: http.MethodGet,
			path:   "/mall/order/:id/transfer",
			wantTrees: &routeTree{
				m: map[string]*node{
					http.MethodGet: {
						typ:  static,
						path: "/",
						children: map[string]*node{
							"mall": {
								typ:  static,
								path: "mall",
								children: map[string]*node{
									"order": {
										typ:  static,
										path: "order",
										paramN: &node{
											typ:  param,
											path: ":id",
											children: map[string]*node{
												"transfer": {
													typ:      static,
													path:     "transfer",
													hdlFunc:  mockHdlFunc,
													children: nil,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}, {
			name:   "regular expression node",
			method: http.MethodGet,
			path:   `/mall/order/re:^\d+$`,
			wantTrees: &routeTree{
				m: map[string]*node{
					http.MethodGet: {
						typ:  static,
						path: "/",
						children: map[string]*node{
							"mall": {
								typ:  static,
								path: "mall",
								children: map[string]*node{
									"order": {
										typ:  static,
										path: "order",
										regexpN: &node{
											typ:     reg,
											re:      regexp.MustCompile(`^\d+$`),
											hdlFunc: mockHdlFunc,
										},
									},
								},
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

	tree.addRoute(http.MethodGet, "/mall/order/*", mockHdlFunc)
	// register wildcard and param node at the same time
	assert.Panics(t, func() {
		tree.addRoute(http.MethodGet, "/mall/order/:id", mockHdlFunc)
	})
	// register wildcard and regexp node at the same time
	assert.Panics(t, func() {
		tree.addRoute(http.MethodGet, "/mall/order/re:^\\d+$", mockHdlFunc)
	})

	// duplicate registered param node
	tree.addRoute(http.MethodGet, "/mall/goods/:id", mockHdlFunc)
	tree.addRoute(http.MethodGet, "/mall/goods/:id/info", mockHdlFunc)
	assert.Panics(t, func() {
		tree.addRoute(http.MethodGet, "/mall/goods/:name", mockHdlFunc)
	})

	// duplicate registered regex node
	tree.addRoute(http.MethodGet, "/mall/items/re:^\\d+$", mockHdlFunc)
	tree.addRoute(http.MethodGet, "/mall/items/re:^\\d+$/details", mockHdlFunc)
	assert.Panics(t, func() {
		tree.addRoute(http.MethodGet, "/mall/items/re:^\\w+$", mockHdlFunc)
	})
}

func TestRouteTree_getRoute(t *testing.T) {
	mockHdlFunc := func(ctx *Context) {}

	tree := newRouteTree()
	tree.addRoute(http.MethodGet, "/", mockHdlFunc)

	tree.addRoute(http.MethodGet, "/v1/user", mockHdlFunc)

	tree.addRoute(http.MethodGet, "/v2/mall/order", mockHdlFunc)
	tree.addRoute(http.MethodGet, "/v2/mall/transaction", mockHdlFunc)
	tree.addRoute(http.MethodGet, "/v2/mall/transaction/:id", mockHdlFunc)
	tree.addRoute(http.MethodGet, "/v2/mall/transaction/:id/customer/:name", mockHdlFunc)

	tree.addRoute(http.MethodPost, "/v2/mall/order", mockHdlFunc)
	tree.addRoute(http.MethodPost, "/v2/mall/transaction", mockHdlFunc)
	tree.addRoute(http.MethodPost, "/v2/mall/transaction/*", mockHdlFunc)
	tree.addRoute(http.MethodPost, "/v2/mall/*/goods", mockHdlFunc)

	tree.addRoute(http.MethodGet, "/v3/mall/oreder/re:^\\d+$", mockHdlFunc)
	tree.addRoute(http.MethodGet, "/v3/email/re:^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$", mockHdlFunc)

	tcs := []struct {
		name     string
		method   string
		path     string
		wantInfo *matched
	}{
		{
			name:   "root node",
			method: http.MethodGet,
			path:   "/",
			wantInfo: &matched{
				ok:      true,
				hdlFunc: mockHdlFunc,
			},
		}, {
			name:     "not found",
			method:   http.MethodGet,
			path:     "/user",
			wantInfo: &matched{ok: false},
		}, {
			name:     "node without hdlFunc",
			method:   http.MethodGet,
			path:     "/v2/mall",
			wantInfo: &matched{ok: false},
		}, {
			name:   "normal",
			method: http.MethodPost,
			path:   "/v2/mall/order",
			wantInfo: &matched{
				ok:      true,
				hdlFunc: mockHdlFunc,
			},
		}, {
			name:   "wildcard node",
			method: http.MethodPost,
			path:   "/v2/mall/transaction/something",
			wantInfo: &matched{
				ok:      true,
				hdlFunc: mockHdlFunc,
			},
		}, {
			name:   "wildcard node multisegment matched",
			method: http.MethodPost,
			path:   "/v2/mall/transaction/a/b/c",
			wantInfo: &matched{
				ok:      true,
				hdlFunc: mockHdlFunc,
			},
		}, {
			name:   "wildcard node multisegment matched in middle",
			method: http.MethodPost,
			path:   "/v2/mall/a/b/c/goods",
			wantInfo: &matched{
				ok:      true,
				hdlFunc: mockHdlFunc,
			},
		}, {
			name:   "param node",
			method: http.MethodGet,
			path:   "/v2/mall/transaction/123",
			wantInfo: &matched{
				ok:      true,
				hdlFunc: mockHdlFunc,
				params:  map[string]string{"id": "123"},
			},
		}, {
			name:   "multiple param nodes",
			method: http.MethodGet,
			path:   "/v2/mall/transaction/123/customer/tom",
			wantInfo: &matched{
				ok:      true,
				hdlFunc: mockHdlFunc,
				params: map[string]string{
					"id":   "123",
					"name": "tom",
				},
			},
		}, {
			name:   "regular exp node matched",
			method: http.MethodGet,
			path:   "/v3/mall/oreder/1234",
			wantInfo: &matched{
				ok:      true,
				hdlFunc: mockHdlFunc,
			},
		}, {
			name:   "regular exp node unmatched",
			method: http.MethodGet,
			path:   "/v3/mall/oreder/abcd",
			wantInfo: &matched{
				ok: false,
			},
		}, {
			name:   "complex regular exp node matched",
			method: http.MethodGet,
			path:   "/v3/email/example@gmail.com",
			wantInfo: &matched{
				ok:      true,
				hdlFunc: mockHdlFunc,
			},
		}, {
			name:   "complex regular exp node unmatched",
			method: http.MethodGet,
			path:   "/v3/email/example@gmail",
			wantInfo: &matched{
				ok: false,
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			mi := tree.getRoute(tc.method, tc.path)
			assert.Equal(t, tc.wantInfo.ok, mi.ok)

			if mi.ok {
				assert.True(t, tc.wantInfo.hdlFunc.equal(mi.hdlFunc))
			}

			tree.putMatchInfo(mi)
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

	if n.wildcardN != nil {
		if other.wildcardN == nil {
			return "wildcardNode not found in other", false
		}

		msg, ok := n.wildcardN.equal(other.wildcardN)
		if !ok {
			return msg, false
		}
	}

	if n.paramN != nil {
		if other.paramN == nil {
			return "paramNode not found in other", false
		}

		msg, ok := n.paramN.equal(other.paramN)
		if !ok {
			return msg, false
		}
	}

	if n.regexpN != nil {
		if other.regexpN == nil {
			return "regexpNode not found in other", false
		}

		if n.regexpN.re.String() != other.regexpN.re.String() {
			return fmt.Sprintf("regexp: %s != %s", n.regexpN.re.String(), other.regexpN.re.String()), false
		}
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

package easyweb

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
						typ:        static,
						baseRoute:  "",
						fullRoute:  "/",
						handleFunc: mockHdlFunc,
						children:   nil,
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
						typ: static,
						children: map[string]*node{
							"user": {
								typ:       static,
								baseRoute: "user",
								children: map[string]*node{
									"test": {
										typ:        static,
										baseRoute:  "test",
										fullRoute:  "/user/test",
										handleFunc: mockHdlFunc,
										children:   nil,
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
						typ: static,
						children: map[string]*node{
							"user": {
								typ:       static,
								baseRoute: "user",
								children: map[string]*node{
									"test": {
										typ:        static,
										baseRoute:  "test",
										fullRoute:  "/user/test",
										handleFunc: mockHdlFunc,
										children:   nil,
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
						typ: static,
						children: map[string]*node{
							"user": {
								typ:        static,
								baseRoute:  "user",
								fullRoute:  "/user",
								handleFunc: mockHdlFunc,
								children:   nil,
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
						typ: static,
						children: map[string]*node{
							"mall": {
								typ:       static,
								baseRoute: "mall",
								children:  map[string]*node{},
								wildcardN: &node{
									typ:        wildcard,
									baseRoute:  "*",
									fullRoute:  "/mall/*",
									handleFunc: mockHdlFunc,
									children:   nil,
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
						typ: static,
						children: map[string]*node{
							"mall": {
								typ:       static,
								baseRoute: "mall",
								children:  map[string]*node{},
								wildcardN: &node{
									typ:       wildcard,
									baseRoute: "*",
									children: map[string]*node{
										"transfer": {
											typ:        static,
											baseRoute:  "transfer",
											fullRoute:  "/mall/*/transfer",
											handleFunc: mockHdlFunc,
											children:   nil,
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
						typ: static,
						children: map[string]*node{
							"mall": {
								typ:       static,
								baseRoute: "mall",
								children: map[string]*node{
									"order": {
										typ:       static,
										baseRoute: "order",
										paramN: &node{
											typ:        param,
											baseRoute:  ":id",
											fullRoute:  "/mall/order/:id",
											handleFunc: mockHdlFunc,
											children:   nil,
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
						typ: static,
						children: map[string]*node{
							"mall": {
								typ:       static,
								baseRoute: "mall",
								children: map[string]*node{
									"order": {
										typ:       static,
										baseRoute: "order",
										paramN: &node{
											typ:       param,
											baseRoute: ":id",
											children: map[string]*node{
												"transfer": {
													typ:        static,
													baseRoute:  "transfer",
													fullRoute:  "/mall/order/:id/transfer",
													handleFunc: mockHdlFunc,
													children:   nil,
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
						typ: static,
						children: map[string]*node{
							"mall": {
								typ:       static,
								baseRoute: "mall",
								children: map[string]*node{
									"order": {
										typ:       static,
										baseRoute: "order",
										regexpN: &node{
											typ:        reg,
											baseRoute:  "re:^\\d+$",
											fullRoute:  `/mall/order/re:^\d+$`,
											re:         regexp.MustCompile(`^\d+$`),
											handleFunc: mockHdlFunc,
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
}

func TestRouteTree_addRoute_panic(t *testing.T) {
	mockHdlFunc := func(ctx *Context) {}

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

func TestRouteTree_addRoute_middleware(t *testing.T) {
	mockHdlFunc := func(ctx *Context) {}
	firstMockMwFunc := func(next HandleFunc) HandleFunc {
		return func(ctx *Context) {
			println("first middleware")
			next(ctx)
		}
	}
	secondMockMwFunc := func(next HandleFunc) HandleFunc {
		return func(ctx *Context) {
			println("second middleware")
			next(ctx)
		}
	}

	tree := newRouteTree()
	tree.addRoute(http.MethodGet, "/mall/goods/:id", mockHdlFunc, firstMockMwFunc, secondMockMwFunc)

	wantTrees := &routeTree{
		m: map[string]*node{
			http.MethodGet: {
				typ: static,
				children: map[string]*node{
					"mall": {
						typ:       static,
						baseRoute: "mall",
						children: map[string]*node{
							"goods": {
								typ:       static,
								baseRoute: "goods",
								paramN: &node{
									typ:             param,
									baseRoute:       ":id",
									fullRoute:       "/mall/goods/:id",
									handleFunc:      mockHdlFunc,
									middlewareChain: MiddlewareChain{firstMockMwFunc, secondMockMwFunc},
									children:        nil,
								},
							},
						},
					},
				},
			},
		},
	}

	msg, ok := tree.equal(wantTrees)
	if !ok {
		t.Log(msg)
	}
	assert.True(t, ok)
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
				node: &node{
					handleFunc: mockHdlFunc,
				},
			},
		}, {
			name:   "not found",
			method: http.MethodGet,
			path:   "/user",
			wantInfo: &matched{
				node: nil,
			},
		}, {
			name:   "node without hdlFunc",
			method: http.MethodGet,
			path:   "/v2/mall",
			wantInfo: &matched{
				node: nil,
			},
		}, {
			name:   "normal",
			method: http.MethodPost,
			path:   "/v2/mall/order",
			wantInfo: &matched{
				node: &node{
					handleFunc: mockHdlFunc,
				},
			},
		}, {
			name:   "wildcard node",
			method: http.MethodPost,
			path:   "/v2/mall/transaction/something",
			wantInfo: &matched{
				node: &node{
					handleFunc: mockHdlFunc,
				},
			},
		}, {
			name:   "wildcard node multisegment matched",
			method: http.MethodPost,
			path:   "/v2/mall/transaction/a/b/c",
			wantInfo: &matched{
				node: &node{
					handleFunc: mockHdlFunc,
				},
			},
		}, {
			name:   "wildcard node multisegment matched in middle",
			method: http.MethodPost,
			path:   "/v2/mall/a/b/c/goods",
			wantInfo: &matched{
				node: &node{
					handleFunc: mockHdlFunc,
				},
			},
		}, {
			name:   "param node",
			method: http.MethodGet,
			path:   "/v2/mall/transaction/123",
			wantInfo: &matched{
				node: &node{
					handleFunc: mockHdlFunc,
				},
				params: map[string]string{
					"id":   "123",
					"name": "tom",
				},
			},
		}, {
			name:   "multiple param nodes",
			method: http.MethodGet,
			path:   "/v2/mall/transaction/123/customer/tom",
			wantInfo: &matched{
				node: &node{
					handleFunc: mockHdlFunc,
				},
			},
		}, {
			name:   "regular exp node matched",
			method: http.MethodGet,
			path:   "/v3/mall/oreder/1234",
			wantInfo: &matched{
				node: &node{
					handleFunc: mockHdlFunc,
				},
			},
		}, {
			name:   "regular exp node unmatched",
			method: http.MethodGet,
			path:   "/v3/mall/oreder/abcd",
			wantInfo: &matched{
				node: nil,
			},
		}, {
			name:   "complex regular exp node matched",
			method: http.MethodGet,
			path:   "/v3/email/example@gmail.com",
			wantInfo: &matched{
				node: &node{
					handleFunc: mockHdlFunc,
				},
			},
		}, {
			name:   "complex regular exp node unmatched",
			method: http.MethodGet,
			path:   "/v3/email/example@gmail",
			wantInfo: &matched{
				node: nil,
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			m := tree.getRoute(tc.method, tc.path)

			if tc.wantInfo.node != nil {
				assert.True(t, tc.wantInfo.node.handleFunc.equal(m.node.handleFunc))
			}

			tree.putMatchInfo(m)
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
	if n.baseRoute != other.baseRoute {
		return fmt.Sprintf("path: %s != %s", n.baseRoute, other.baseRoute), false
	}

	if n.fullRoute != other.fullRoute {
		return fmt.Sprintf("fullRoute: %s != %s", n.fullRoute, other.fullRoute), false
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

	if n.handleFunc != nil {
		if other.handleFunc == nil {
			return "hdlFunc not found in other", false
		}

		if !n.handleFunc.equal(other.handleFunc) {
			return "hdlFunc not equal", false
		}
	}

	if n.middlewareChain != nil {
		if other.middlewareChain == nil {
			return "mwChain not found in other", false
		}

		if !n.middlewareChain.equal(other.middlewareChain) {
			return "mwChain not equal", false
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

func (h HandleFunc) equal(other HandleFunc) bool {
	return reflect.ValueOf(h) == reflect.ValueOf(other)
}

func (m MiddlewareChain) equal(other MiddlewareChain) bool {
	if len(m) != len(other) {
		return false
	}

	for i, mw := range m {
		if reflect.ValueOf(mw) != reflect.ValueOf(other[i]) {
			return false
		}
	}
	return true
}

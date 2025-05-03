package easyweb

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRouteGroup_Get(t *testing.T) {
	svr := NewHttpServer()
	rg := svr.Group("/api")
	rg.Route(http.MethodGet, "/user", func(ctx *Context) {})

	mi := svr.getRoute(http.MethodGet, "/api/user")
	assert.NotNil(t, mi.node)
	assert.NotNil(t, mi.node.handleFunc)
}

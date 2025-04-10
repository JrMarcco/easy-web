package easy_web

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRouteGroup_Get(t *testing.T) {
	svr := NewHttpSvr()
	rg := svr.Group("/api")
	rg.Get("/user", func(ctx *Context) {})

	route, ok := svr.getRoute(http.MethodGet, "/api/user")
	assert.True(t, ok)
	assert.NotNil(t, route.hdlFunc)
	assert.Equal(t, "user", route.path)
}

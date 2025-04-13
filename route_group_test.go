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

	mi := svr.getRoute(http.MethodGet, "/api/user")
	assert.True(t, mi.ok)
	assert.NotNil(t, mi.hdlFunc)
}

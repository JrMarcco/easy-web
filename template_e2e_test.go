//go:build e2e

package easyweb

import (
	"github.com/stretchr/testify/require"
	"html/template"
	"log"
	"net/http"
	"testing"
)

func TestGoTemplateEngine_Render(t *testing.T) {
	tpl, err := template.ParseGlob("testdata/templates/*.gohtml")
	require.NoError(t, err)

	srv := NewHttpServer(ServerWithTplEngineOpt(&GoTemplateEngine{T: tpl}))

	srv.Route(http.MethodGet, "/login", func(ctx *Context) {
		err := ctx.Render("login.gohtml", nil)
		if err != nil {
			log.Println(err)
		}
	})

	err = srv.Start()
	if err != nil {
		t.Fatal(err)
	}
}

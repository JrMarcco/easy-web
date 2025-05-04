//go:build e2e

package easyweb

import (
	"github.com/stretchr/testify/require"
	"html/template"
	"log"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"testing"
)

func TestFileUploader_Handle(t *testing.T) {
	tpl, err := template.ParseGlob("testdata/templates/*.gohtml")
	require.NoError(t, err)

	srv := NewHttpServer(ServerWithTplEngineOpt(&GoTemplateEngine{T: tpl}))

	srv.Route(http.MethodGet, "/test", func(ctx *Context) {
		err := ctx.Render("upload.gohtml", nil)
		if err != nil {
			log.Println(err)
		}
	})

	fu := FileUploader{
		FileField: "file",
		DstPathFunc: func(mph *multipart.FileHeader) string {
			return filepath.Join("testdata", "upload", mph.Filename)
		},
	}

	srv.Route(http.MethodPost, "/upload", fu.Handle())

	err = srv.Start()
}

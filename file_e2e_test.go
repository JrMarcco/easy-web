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

	srv.Route(
		http.MethodPost,
		"/upload",
		NewFileUploader(
			FileUploaderWithFieldName("file"),
			FileUploaderWithDstPathFunc(func(mph *multipart.FileHeader) string {
				return filepath.Join("testdata", "upload", mph.Filename)
			}),
		).Handle(),
	)

	err = srv.Start()
	if err != nil {
		t.Fatal(err)
	}
}

func TestFileDownloader_Handle(t *testing.T) {
	srv := NewHttpServer()

	srv.Route(
		http.MethodGet,
		"/download",
		NewFileDownloader(
			FileDownloaderWithFilePath("testdata/upload"),
			FileDownloaderWithFieldName("file"),
		).Handle(),
	)

	err := srv.Start()
	if err != nil {
		t.Fatal(err)
	}
}

func TestStaticFileServer_Handle(t *testing.T) {
	srv := NewHttpServer()

	srh := NewStaticResourceHandler(
		StaticResourceHandlerWithFieldName("file"),
		StaticResourceHandlerWithFilePath("testdata/statics"),
		StaticResourceHandlerWithCache(128, 1024*1024),
	)

	srv.Route(http.MethodGet, "/static/:file", srh.Handle())

	err := srv.Start()
	if err != nil {
		t.Fatal(err)
	}
}

//go:build e2e

package prometheus

import (
	"math/rand"
	"net/http"
	"testing"
	"time"

	easyweb "github.com/JrMarcco/easy-web"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func TestMiddlewareBuilder_Build(t *testing.T) {
	svr := easyweb.NewHttpServer()

	svr.Route(http.MethodGet, "/prometheus/test", func(ctx *easyweb.Context) {
		val := rand.Intn(1000) + 1
		time.Sleep(time.Millisecond * time.Duration(val))

		_ = ctx.Ok()
	}, NewMiddlewareBuilder().Build())

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		_ = http.ListenAndServe(":8081", nil)
	}()

	_ = svr.Start()
}

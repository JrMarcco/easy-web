package test

import (
	"github.com/jrmarcco/go-learning/web"
	"github.com/jrmarcco/go-learning/web/middleware/accesslog"
	"github.com/jrmarcco/go-learning/web/middleware/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"math/rand"
	"net/http"
	"testing"
	"time"
)

func TestWeb_HttpServer(t *testing.T) {

	h := web.NewHttpServer(web.SvrWithAddr(":8081"))

	userApi := h.Group("/user")

	pb := &prometheus.MiddlewareBuilder{
		Name:      "jrmarcco",
		Subsystem: "http_request",
		ConstLabels: map[string]string{
			"inst_id": "inst_8888",
		},
		Objectives: map[float64]float64{
			0.5:   0.01,
			0.75:  0.01,
			0.90:  0.01,
			0.99:  0.001,
			0.999: 0.0001,
		},
	}

	userApi.Use(
		pb.Builder(),
		accesslog.NewBuilder().Build(),
	)

	{
		userApi.Get("/list", func(ctx *web.Context) {
			time.Sleep(
				time.Duration(rand.Intn(1000)+1) * time.Millisecond,
			)
			if err := ctx.OkJson("hello"); err != nil {
				log.Fatalln(err)
			}
		})
	}

	go func() {
		http.Handle("/metrics", promhttp.Handler())

		err := http.ListenAndServe(":2112", nil)
		if err != nil {
			log.Fatalln(err)
		}
	}()

	if err := h.Start(); err != nil {
		log.Fatalln(err)
	}

}

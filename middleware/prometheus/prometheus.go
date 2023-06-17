package prometheus

import (
	"github.com/jrmarcco/go-learning/web"
	"github.com/prometheus/client_golang/prometheus"
	"strconv"
	"time"
)

type MiddlewareBuilder struct {
	Name        string
	Subsystem   string
	ConstLabels map[string]string
	Objectives  map[float64]float64
	Help        string
}

func (m *MiddlewareBuilder) Builder() web.Middleware {

	summaryVec := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name:        m.Name,
		Subsystem:   m.Subsystem,
		ConstLabels: m.ConstLabels,
		Objectives:  m.Objectives,
		Help:        m.Help,
	}, []string{"pattern", "method", "status"})

	prometheus.MustRegister(summaryVec)

	return func(next web.HandleFunc) web.HandleFunc {
		return func(ctx *web.Context) {

			startTime := time.Now()
			next(ctx)
			endTime := time.Now()

			go report(endTime.Sub(startTime), ctx, summaryVec)
		}
	}
}

func report(duration time.Duration, ctx *web.Context, vec prometheus.ObserverVec) {
	route := "unknown"
	if ctx.MatchedRoute != "" {
		route = ctx.MatchedRoute
	}

	vec.WithLabelValues(
		route,
		ctx.Req.Method,
		strconv.Itoa(ctx.RspStatusCode),
	).Observe(float64(duration.Milliseconds()))
}

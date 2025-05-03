package prometheus

import (
	easyweb "github.com/JrMarcco/easy-web"
	"github.com/prometheus/client_golang/prometheus"
	"strconv"
	"time"
)

type MiddlewareBuilder struct {
	vec *prometheus.SummaryVec
}

func (b *MiddlewareBuilder) WithSummaryVec(vec *prometheus.SummaryVec) *MiddlewareBuilder {
	b.vec = vec
	return b
}

func (b *MiddlewareBuilder) Build() easyweb.Middleware {
	if b.vec == nil {
		b.vec = prometheus.NewSummaryVec(
			prometheus.SummaryOpts{
				Name:        "easy_web",
				Subsystem:   "http_request_duration",
				ConstLabels: map[string]string{},
				Objectives: map[float64]float64{
					0.5:   0.01,
					0.75:  0.01,
					0.90:  0.005,
					0.98:  0.002,
					0.99:  0.001,
					0.999: 0.0001,
				},
				Help: "Duration of HTTP requests in microseconds.",
			},
			[]string{"method", "path", "status_code"},
		)
	}

	prometheus.MustRegister(b.vec)

	return func(next easyweb.HandleFunc) easyweb.HandleFunc {
		return func(ctx *easyweb.Context) {
			start := time.Now()
			next(ctx)
			end := time.Now()

			go b.report(ctx, end.Sub(start))
		}
	}
}

func (b *MiddlewareBuilder) report(ctx *easyweb.Context, duration time.Duration) {
	path := "unknown"
	if ctx.MatchedRoute != "" {
		path = ctx.MatchedRoute
	}

	b.vec.WithLabelValues(
		ctx.Req.Method,
		path,
		strconv.Itoa(ctx.StatusCode),
	).Observe(float64(duration.Microseconds()))
}

func NewMiddlewareBuilder() *MiddlewareBuilder {
	return &MiddlewareBuilder{}
}

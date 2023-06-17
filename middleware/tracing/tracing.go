package tracing

import (
	"github.com/jrmarcco/go-learning/web"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const defaultInstrumentationName = "github.com/jrmarcco/go-learning/web/middleware/otel"

type MiddlewareBuilder struct {
	tracer trace.Tracer
}

func NewBuilder() *MiddlewareBuilder {
	tracer := otel.GetTracerProvider().Tracer(defaultInstrumentationName)

	return &MiddlewareBuilder{
		tracer: tracer,
	}
}

func (m *MiddlewareBuilder) Build() web.Middleware {
	return func(next web.HandleFunc) web.HandleFunc {
		return func(ctx *web.Context) {

			extractCtx := otel.GetTextMapPropagator().Extract(ctx.TraceCtx, propagation.HeaderCarrier(ctx.Req.Header))
			extractCtx, span := m.tracer.Start(extractCtx, "unknown", trace.WithAttributes())

			span.SetAttributes(
				attribute.String("http.method", ctx.Req.Method),
				attribute.String("http.proto", ctx.Req.Proto),
				attribute.String("http.host", ctx.Req.Host),
				attribute.String("http.url", ctx.Req.URL.String()),
			)

			defer span.End()

			ctx.TraceCtx = extractCtx
			next(ctx)

			if ctx.MatchedRoute != "" {
				span.SetName(ctx.MatchedRoute)
			}

			span.SetAttributes(attribute.Int("http.status", ctx.RspStatusCode))
		}
	}
}

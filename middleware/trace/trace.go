package trace

import (
	easyweb "github.com/JrMarcco/easy-web"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const defaultInstrumentationName = "github.com/JrMarcco/easy-web"

type MiddlewareBuilder struct {
	tracer trace.Tracer
}

func (b *MiddlewareBuilder) Build() easyweb.Middleware {
	return func(next easyweb.HandleFunc) easyweb.HandleFunc {
		return func(ctx *easyweb.Context) {
			extractCtx := otel.GetTextMapPropagator().Extract(ctx.TraceCtx, propagation.HeaderCarrier(ctx.Req.Header))
			extractCtx, span := b.tracer.Start(extractCtx, "unknow", trace.WithAttributes())

			span.SetAttributes(
				attribute.String("http.method", ctx.Req.Method),
				attribute.String("http.proto", ctx.Req.Proto),
				attribute.String("http.host", ctx.Req.Host),
				attribute.String("http.path", ctx.Req.URL.Path),
			)

			defer span.End()

			ctx.TraceCtx = extractCtx
			next(ctx)

			if ctx.MatchedRoute != "" {
				span.SetName(ctx.MatchedRoute)
			}

			span.SetAttributes(attribute.Int("http.status_code", ctx.StatusCode))
		}
	}
}

func NewMiddlewareBuilder() *MiddlewareBuilder {
	tracer := otel.GetTracerProvider().Tracer(defaultInstrumentationName)
	return &MiddlewareBuilder{tracer: tracer}
}

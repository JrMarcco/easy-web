package trace

import (
	easyweb "github.com/JrMarcco/easy-web"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const defaultInstrumentationName = "github.com/JrMarcco/easy-web/middleware/trace"

type MiddlewareBuilder struct {
	Tracer trace.Tracer
}

// WithTracer customizes the tracer.
func (b *MiddlewareBuilder) WithTracer(tracer trace.Tracer) *MiddlewareBuilder {
	b.Tracer = tracer
	return b
}

func (b *MiddlewareBuilder) Build() easyweb.Middleware {
	// if the user did not provide a tracer,
	// build a default opentelemetry tracer
	if b.Tracer == nil {
		b.Tracer = otel.GetTracerProvider().Tracer(defaultInstrumentationName)
	}

	return func(next easyweb.HandleFunc) easyweb.HandleFunc {
		return func(ctx *easyweb.Context) {
			extractCtx := otel.GetTextMapPropagator().Extract(ctx.TraceCtx, propagation.HeaderCarrier(ctx.Req.Header))
			extractCtx, span := b.Tracer.Start(extractCtx, "unknow", trace.WithAttributes())

			span.SetAttributes(
				attribute.String("http.method", ctx.Req.Method),
				attribute.String("http.proto", ctx.Req.Proto),
				attribute.String("http.host", ctx.Req.Host),
				attribute.String("http.scheme", ctx.Req.URL.Scheme),
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
	return &MiddlewareBuilder{}
}

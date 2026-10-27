package telemetry

import (
	"context"

	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const inboundInstrumentation = "github.com/Onurlulardan/harpy/platform/http"

func (runtime *Runtime) Extract(
	ctx context.Context,
	carrier propagation.TextMapCarrier,
) context.Context {
	return runtime.propagator.Extract(ctx, carrier)
}

func (runtime *Runtime) Inject(
	ctx context.Context,
	carrier propagation.TextMapCarrier,
) {
	runtime.propagator.Inject(ctx, carrier)
}

func (runtime *Runtime) StartInbound(
	ctx context.Context,
	carrier propagation.TextMapCarrier,
	spanName string,
	options ...trace.SpanStartOption,
) (context.Context, trace.Span) {
	parent := runtime.Extract(ctx, carrier)
	options = append(options, trace.WithSpanKind(trace.SpanKindServer))
	return runtime.tracerProvider.Tracer(inboundInstrumentation).Start(parent, spanName, options...)
}

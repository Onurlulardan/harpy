package telemetry

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"go.opentelemetry.io/otel/baggage"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

const (
	testTraceID = "0af7651916cd43dd8448eb211c80319c"
	testSpanID  = "b7ad6b7169203331"
)

func TestExtractAcceptsOnlyValidTraceContext(t *testing.T) {
	runtime := propagationRuntime(t)
	carrier := propagation.MapCarrier{
		"traceparent": "00-" + testTraceID + "-" + testSpanID + "-01",
		"tracestate":  "vendor=value",
		"baggage":     "credential=must-not-propagate",
	}
	extracted := runtime.Extract(context.Background(), carrier)
	spanContext := trace.SpanContextFromContext(extracted)
	if !spanContext.IsValid() || !spanContext.IsRemote() || !spanContext.IsSampled() {
		t.Fatalf("valid remote trace context was not preserved: %#v", spanContext)
	}
	if spanContext.TraceState().String() != "vendor=value" {
		t.Fatalf("unexpected tracestate: %s", spanContext.TraceState())
	}
	if baggage.FromContext(extracted).Len() != 0 {
		t.Fatal("baggage header gained authority")
	}
}

func TestExtractRejectsMalformedAndAllZeroParents(t *testing.T) {
	runtime := propagationRuntime(t)
	parents := []string{
		"00-00000000000000000000000000000000-" + testSpanID + "-01",
		"00-" + testTraceID + "-0000000000000000-01",
		"ff-" + testTraceID + "-" + testSpanID + "-01",
		"not-a-trace-parent",
	}
	for _, parent := range parents {
		t.Run(parent, func(t *testing.T) {
			carrier := propagation.MapCarrier{
				"traceparent": parent,
				"tracestate":  "vendor=value",
			}
			if got := trace.SpanContextFromContext(runtime.Extract(context.Background(), carrier)); got.IsValid() {
				t.Fatalf("invalid parent accepted: %s", parent)
			}
		})
	}
	stateOnly := propagation.MapCarrier{"tracestate": "vendor=value"}
	if got := trace.SpanContextFromContext(runtime.Extract(context.Background(), stateOnly)); got.IsValid() {
		t.Fatal("tracestate without traceparent was accepted")
	}
}

func TestInjectDoesNotWriteBaggage(t *testing.T) {
	runtime := propagationRuntime(t)
	traceID, _ := trace.TraceIDFromHex(testTraceID)
	spanID, _ := trace.SpanIDFromHex(testSpanID)
	spanContext := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID: traceID, SpanID: spanID, TraceFlags: trace.FlagsSampled,
	})
	member, _ := baggage.NewMember("credential", "must-not-propagate")
	bag, _ := baggage.New(member)
	ctx := baggage.ContextWithBaggage(trace.ContextWithSpanContext(context.Background(), spanContext), bag)
	carrier := propagation.MapCarrier{}
	runtime.Inject(ctx, carrier)
	if carrier.Get("traceparent") == "" || carrier.Get("baggage") != "" {
		t.Fatalf("unexpected injected carrier: %#v", carrier)
	}
}

func TestStartInboundCreatesServerSpanWithRemoteParent(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	provider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithSyncer(exporter),
	)
	failures := newFailureReporter(slog.New(slog.NewJSONHandler(io.Discard, nil)))
	runtime := assembleRuntime(
		provider, metricnoop.NewMeterProvider(), resource.Empty(), failures,
		provider.Shutdown, func(context.Context) error { return nil },
	)
	carrier := propagation.MapCarrier{
		"traceparent": "00-" + testTraceID + "-" + testSpanID + "-01",
	}
	_, span := runtime.StartInbound(
		context.Background(), carrier, "GET /forms", trace.WithSpanKind(trace.SpanKindClient),
	)
	span.End()
	spans := exporter.GetSpans()
	if len(spans) != 1 || spans[0].SpanKind != trace.SpanKindServer {
		t.Fatalf("unexpected inbound spans: %#v", spans)
	}
	if spans[0].Parent.SpanID().String() != testSpanID || !spans[0].Parent.IsRemote() {
		t.Fatalf("remote parent was not retained: %#v", spans[0].Parent)
	}
	if err := runtime.Shutdown(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func propagationRuntime(t *testing.T) *Runtime {
	t.Helper()
	runtime, err := New(
		context.Background(), telemetrySettings(t, false), telemetryBuildInfo(t),
		slog.New(slog.NewJSONHandler(io.Discard, nil)),
	)
	if err != nil {
		t.Fatal(err)
	}
	return runtime
}

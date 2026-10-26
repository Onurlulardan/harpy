package telemetry

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"go.opentelemetry.io/otel"
)

func TestDisabledRuntimeDoesNotMutateGlobalProviders(t *testing.T) {
	globalTrace := otel.GetTracerProvider()
	globalMetric := otel.GetMeterProvider()
	runtime, err := New(
		context.Background(), telemetrySettings(t, false), telemetryBuildInfo(t),
		slog.New(slog.NewJSONHandler(io.Discard, nil)),
	)
	if err != nil {
		t.Fatal(err)
	}
	if otel.GetTracerProvider() != globalTrace || otel.GetMeterProvider() != globalMetric {
		t.Fatal("telemetry bootstrap mutated a process-global provider")
	}
	if runtime.TracerProvider() == nil || runtime.MeterProvider() == nil {
		t.Fatal("disabled runtime did not provide no-export providers")
	}
	if runtime.resource == nil || runtime.ExportFailures() != (FailureCounts{}) {
		t.Fatal("disabled runtime is incomplete")
	}
	if err := runtime.Shutdown(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestBootstrapRequiresContextAndStructuredLogger(t *testing.T) {
	settings := telemetrySettings(t, false)
	info := telemetryBuildInfo(t)
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	if _, err := New(nil, settings, info, logger); err == nil {
		t.Fatal("nil bootstrap context accepted")
	}
	if _, err := New(context.Background(), settings, info, nil); err == nil {
		t.Fatal("nil structured logger accepted")
	}
}

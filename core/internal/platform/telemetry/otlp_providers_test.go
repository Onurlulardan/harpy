package telemetry

import (
	"context"
	"strings"
	"testing"

	"go.opentelemetry.io/otel"
)

func TestProviderRuntimeExportsAndShutsDownWithoutGlobals(t *testing.T) {
	settings := telemetrySettings(t, true)
	res, err := newResource(telemetryBuildInfo(t), "00000000-0000-4000-8000-000000000000")
	if err != nil {
		t.Fatal(err)
	}
	spanExporter := &fakeSpanExporter{}
	metricExporter := &fakeMetricExporter{}
	failures := newFailureReporter(testLogger(&lockedBuffer{}))
	globalTrace := otel.GetTracerProvider()
	globalMetric := otel.GetMeterProvider()
	runtime, err := newProviderRuntime(
		context.Background(), settings, res, failures, spanExporter, metricExporter,
	)
	if err != nil {
		t.Fatal(err)
	}
	_, span := runtime.TracerProvider().Tracer("test/integration").Start(context.Background(), "operation")
	span.End()
	counter, err := runtime.MeterProvider().Meter("test/integration").Int64Counter("test.operations")
	if err != nil {
		t.Fatal(err)
	}
	counter.Add(context.Background(), 1)
	if err := runtime.Shutdown(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Shutdown(context.Background()); err != nil {
		t.Fatal(err)
	}
	spanCount, spanShutdowns := spanExporter.snapshot()
	metricExports, _, metricShutdowns := metricExporter.snapshot()
	if spanCount != 1 || spanShutdowns != 1 {
		t.Fatalf("trace lifecycle: spans=%d shutdowns=%d", spanCount, spanShutdowns)
	}
	if metricExports == 0 || metricShutdowns != 1 {
		t.Fatalf("metric lifecycle: exports=%d shutdowns=%d", metricExports, metricShutdowns)
	}
	if otel.GetTracerProvider() != globalTrace || otel.GetMeterProvider() != globalMetric {
		t.Fatal("provider runtime mutated process-global providers")
	}
}

func TestExportFailureDoesNotBreakShutdownOrLeakSecret(t *testing.T) {
	settings := telemetrySettings(t, true)
	res, err := newResource(telemetryBuildInfo(t), "00000000-0000-4000-8000-000000000000")
	if err != nil {
		t.Fatal(err)
	}
	output := &lockedBuffer{}
	failures := newFailureReporter(testLogger(output))
	runtime, err := newProviderRuntime(
		context.Background(), settings, res, failures,
		&fakeSpanExporter{exportErr: secretFailure()},
		&fakeMetricExporter{exportErr: secretFailure()},
	)
	if err != nil {
		t.Fatal(err)
	}
	_, span := runtime.TracerProvider().Tracer("test/failure").Start(context.Background(), "failure")
	span.End()
	if err := runtime.Shutdown(context.Background()); err != nil {
		t.Fatalf("export availability affected shutdown: %v", err)
	}
	counts := runtime.ExportFailures()
	if counts.Trace == 0 || counts.Metric == 0 {
		t.Fatalf("export failures were not observable: %#v", counts)
	}
	if containsSecret(output.String()) {
		t.Fatalf("export failure leaked a secret: %s", output.String())
	}
}

func containsSecret(value string) bool {
	return strings.Contains(value, "secret-token") || strings.Contains(value, "bearer")
}

package telemetry

import (
	"context"
	"strings"
	"testing"

	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

func TestExporterFailuresAreCountedLoggedAndSwallowed(t *testing.T) {
	output := &lockedBuffer{}
	reporter := newFailureReporter(testLogger(output))
	spanDelegate := &fakeSpanExporter{exportErr: secretFailure()}
	metricDelegate := &fakeMetricExporter{exportErr: secretFailure()}
	spanExporter := observedSpanExporter{delegate: spanDelegate, failures: reporter}
	metricExporter := observedMetricExporter{delegate: metricDelegate, failures: reporter}
	if err := spanExporter.ExportSpans(context.Background(), nil); err != nil {
		t.Fatalf("trace export failure escaped: %v", err)
	}
	if err := metricExporter.Export(context.Background(), &metricdata.ResourceMetrics{}); err != nil {
		t.Fatalf("metric export failure escaped: %v", err)
	}
	if strings.Contains(output.String(), "secret-token") {
		t.Fatalf("exporter error leaked to structured log: %s", output.String())
	}
	if !strings.Contains(output.String(), `"telemetry.signal":"trace"`) ||
		!strings.Contains(output.String(), `"telemetry.signal":"metric"`) {
		t.Fatalf("structured signal fields missing: %s", output.String())
	}
	if counts := reporter.snapshot(); counts.Trace != 1 || counts.Metric != 1 {
		t.Fatalf("unexpected exporter failure counts: %#v", counts)
	}
}

func TestExporterLifecycleReturnsOnlySanitizedErrors(t *testing.T) {
	reporter := newFailureReporter(testLogger(&lockedBuffer{}))
	spanExporter := observedSpanExporter{
		delegate: &fakeSpanExporter{shutdownErr: secretFailure()}, failures: reporter,
	}
	metricExporter := observedMetricExporter{
		delegate: &fakeMetricExporter{flushErr: secretFailure(), shutdownErr: secretFailure()},
		failures: reporter,
	}
	for name, err := range map[string]error{
		"trace shutdown":  spanExporter.Shutdown(context.Background()),
		"metric flush":    metricExporter.ForceFlush(context.Background()),
		"metric shutdown": metricExporter.Shutdown(context.Background()),
	} {
		if err == nil || strings.Contains(err.Error(), "secret-token") {
			t.Fatalf("%s returned unsafe result: %v", name, err)
		}
	}
}

func TestFailureCounterIsCollectable(t *testing.T) {
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	reporter := newFailureReporter(testLogger(&lockedBuffer{}))
	if err := reporter.bind(provider); err != nil {
		t.Fatal(err)
	}
	reporter.report(context.Background(), signalTrace, operationExport)
	var metrics metricdata.ResourceMetrics
	if err := reader.Collect(context.Background(), &metrics); err != nil {
		t.Fatal(err)
	}
	if got := collectedFailureCount(metrics); got != 1 {
		t.Fatalf("collected failure count: got %d, want 1", got)
	}
	if err := provider.Shutdown(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func collectedFailureCount(metrics metricdata.ResourceMetrics) int64 {
	for _, scope := range metrics.ScopeMetrics {
		for _, current := range scope.Metrics {
			if current.Name != "harpy.telemetry.export.failures" {
				continue
			}
			sum, ok := current.Data.(metricdata.Sum[int64])
			if ok && len(sum.DataPoints) == 1 {
				return sum.DataPoints[0].Value
			}
		}
	}
	return 0
}

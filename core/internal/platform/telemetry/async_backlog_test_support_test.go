package telemetry

import (
	"context"
	"testing"

	asyncobservability "github.com/Onurlulardan/harpy/core/internal/platform/async/observability"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

type asyncSnapshotSource func(context.Context) (asyncobservability.Snapshot, error)

func (source asyncSnapshotSource) Snapshot(
	ctx context.Context,
) (asyncobservability.Snapshot, error) {
	return source(ctx)
}

type asyncMetricCollector struct {
	reader   *sdkmetric.ManualReader
	provider *sdkmetric.MeterProvider
}

func newAsyncMetricCollector(t *testing.T) asyncMetricCollector {
	t.Helper()
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	t.Cleanup(func() {
		if err := provider.Shutdown(context.Background()); err != nil {
			t.Errorf("shutdown async metric provider: %v", err)
		}
	})
	return asyncMetricCollector{reader: reader, provider: provider}
}

func (collector asyncMetricCollector) collect(t *testing.T) map[string]int64 {
	t.Helper()
	var resourceMetrics metricdata.ResourceMetrics
	if err := collector.reader.Collect(context.Background(), &resourceMetrics); err != nil {
		t.Fatalf("collect async backlog metrics: %v", err)
	}
	values := make(map[string]int64)
	for _, scope := range resourceMetrics.ScopeMetrics {
		for _, current := range scope.Metrics {
			switch data := current.Data.(type) {
			case metricdata.Gauge[int64]:
				assertSingleAsyncPoint(t, current.Name, len(data.DataPoints))
				assertNoAsyncAttributes(t, current.Name, data.DataPoints[0].Attributes.Len())
				values[current.Name] = data.DataPoints[0].Value
			case metricdata.Sum[int64]:
				assertSingleAsyncPoint(t, current.Name, len(data.DataPoints))
				assertNoAsyncAttributes(t, current.Name, data.DataPoints[0].Attributes.Len())
				values[current.Name] = data.DataPoints[0].Value
			}
		}
	}
	return values
}

func assertSingleAsyncPoint(t *testing.T, name string, count int) {
	t.Helper()
	if count != 1 {
		t.Fatalf("async metric %s has %d points, want 1", name, count)
	}
}

func assertNoAsyncAttributes(t *testing.T, name string, count int) {
	t.Helper()
	if count != 0 {
		t.Fatalf("async metric %s has %d attributes, want zero", name, count)
	}
}

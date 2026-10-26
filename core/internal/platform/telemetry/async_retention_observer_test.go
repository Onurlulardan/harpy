package telemetry

import (
	"context"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/async/retention"
	"go.opentelemetry.io/otel/attribute"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

func TestAsyncRetentionObserverEmitsOnlyFixedLowCardinalityMetrics(t *testing.T) {
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	t.Cleanup(func() { _ = provider.Shutdown(context.Background()) })
	observerValue, err := newAsyncRetentionObserver(provider)
	if err != nil {
		t.Fatal(err)
	}
	observer := observerValue.(*asyncRetentionObserver)
	counts := retention.StepCounts{
		JobsRedacted: 1, OutboxRedacted: 2, DeadLettersRedacted: 3,
		InboxPurged: 4, IdempotencyPurged: 5, JobsPurged: 6,
		DeadLettersPurged: 7,
	}
	worked, err := retention.NewCycleObservation(
		retention.OutcomeWorked, 25*time.Millisecond, counts,
	)
	if err != nil {
		t.Fatal(err)
	}
	idle, err := retention.NewCycleObservation(
		retention.OutcomeIdle, 5*time.Millisecond, retention.StepCounts{},
	)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	observer.ObserveCycle(ctx, worked)
	observer.ObserveCycle(ctx, idle)
	observer.ObserveFailure(ctx, 10*time.Millisecond)
	observer.ObserveBacklogYield(ctx)

	var collected metricdata.ResourceMetrics
	if err := reader.Collect(ctx, &collected); err != nil {
		t.Fatal(err)
	}
	assertRetentionSums(t, collected, map[string]int64{
		"harpy.async.retention.cycles":         3,
		"harpy.async.retention.rows":           28,
		"harpy.async.retention.failures":       1,
		"harpy.async.retention.backlog_yields": 1,
	})
	assertRetentionAttributes(t, collected)
}

func TestAsyncRetentionObserverRequiresProductionMeter(t *testing.T) {
	if observer, err := (*Runtime)(nil).NewAsyncRetentionObserver(); err == nil || observer != nil {
		t.Fatalf("nil runtime observer = %#v, %v", observer, err)
	}
	if observer, err := newAsyncRetentionObserver(nil); err == nil || observer != nil {
		t.Fatalf("nil provider observer = %#v, %v", observer, err)
	}
}

func assertRetentionSums(
	t *testing.T,
	metrics metricdata.ResourceMetrics,
	want map[string]int64,
) {
	t.Helper()
	got := make(map[string]int64)
	for _, scope := range metrics.ScopeMetrics {
		for _, current := range scope.Metrics {
			data, ok := current.Data.(metricdata.Sum[int64])
			if !ok {
				continue
			}
			for _, point := range data.DataPoints {
				got[current.Name] += point.Value
			}
		}
	}
	for name, value := range want {
		if got[name] != value {
			t.Fatalf("retention metric %s = %d, want %d", name, got[name], value)
		}
	}
}

func assertRetentionAttributes(t *testing.T, metrics metricdata.ResourceMetrics) {
	t.Helper()
	allowed := map[attribute.Key]bool{"outcome": true, "step": true}
	var durationPoints int
	for _, scope := range metrics.ScopeMetrics {
		for _, current := range scope.Metrics {
			switch data := current.Data.(type) {
			case metricdata.Sum[int64]:
				for _, point := range data.DataPoints {
					assertFixedAttributeSet(t, point.Attributes.ToSlice(), allowed)
				}
			case metricdata.Histogram[float64]:
				durationPoints += len(data.DataPoints)
				for _, point := range data.DataPoints {
					assertFixedAttributeSet(t, point.Attributes.ToSlice(), allowed)
				}
			}
		}
	}
	if durationPoints != 3 {
		t.Fatalf("retention duration points = %d, want 3", durationPoints)
	}
}

func assertFixedAttributeSet(
	t *testing.T,
	attributes []attribute.KeyValue,
	allowed map[attribute.Key]bool,
) {
	t.Helper()
	for _, value := range attributes {
		if !allowed[value.Key] {
			t.Fatalf("unbounded retention attribute = %s", value.Key)
		}
	}
}

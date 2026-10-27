package telemetry

import (
	"context"
	"testing"

	asyncobservability "github.com/Onurlulardan/harpy/core/internal/platform/async/observability"
)

func TestAsyncBacklogMetricsAreCollectableWithoutAttributes(t *testing.T) {
	snapshot, err := asyncobservability.NewSnapshot(301, 1, 72, 2, 3)
	if err != nil {
		t.Fatal(err)
	}
	collector := newAsyncMetricCollector(t)
	runtime := &Runtime{meterProvider: collector.provider}
	source := asyncSnapshotSource(func(context.Context) (asyncobservability.Snapshot, error) {
		return snapshot, nil
	})
	if err := runtime.RegisterAsyncBacklog(source, testLogger(&lockedBuffer{})); err != nil {
		t.Fatal(err)
	}
	want := map[string]int64{
		"harpy.async.outbox.undelivered.oldest_age": 301,
		"harpy.async.outbox.undelivered.alarm":      1,
		"harpy.async.job.due.oldest_lag":            72,
		"harpy.async.job.parked":                    2,
		"harpy.async.outbox.parked":                 3,
		"harpy.async.observer.failures":             0,
	}
	got := collector.collect(t)
	if len(got) != len(want) {
		t.Fatalf("collected async metric count = %d, want %d: %#v", len(got), len(want), got)
	}
	for name, value := range want {
		if got[name] != value {
			t.Fatalf("async metric %s = %d, want %d", name, got[name], value)
		}
	}
}

func TestRegisterAsyncBacklogRejectsIncompleteDependencies(t *testing.T) {
	collector := newAsyncMetricCollector(t)
	source := asyncSnapshotSource(func(context.Context) (asyncobservability.Snapshot, error) {
		return asyncobservability.Snapshot{}, nil
	})
	if err := (*Runtime)(nil).RegisterAsyncBacklog(source, testLogger(&lockedBuffer{})); err == nil {
		t.Fatal("nil telemetry runtime was accepted")
	}
	runtime := &Runtime{meterProvider: collector.provider}
	if err := runtime.RegisterAsyncBacklog(nil, testLogger(&lockedBuffer{})); err == nil {
		t.Fatal("nil async source was accepted")
	}
	if err := runtime.RegisterAsyncBacklog(source, nil); err == nil {
		t.Fatal("nil async logger was accepted")
	}
}

package telemetry

import (
	"context"
	"errors"
	"strings"
	"testing"

	asyncobservability "github.com/Onurlulardan/harpy/core/internal/platform/async/observability"
)

func TestAsyncBacklogObservationFailureIsCountedLoggedAndSwallowed(t *testing.T) {
	collector := newAsyncMetricCollector(t)
	output := &lockedBuffer{}
	source := asyncSnapshotSource(func(context.Context) (asyncobservability.Snapshot, error) {
		return asyncobservability.Snapshot{}, errors.New("database secret-token unavailable")
	})
	if err := registerAsyncBacklog(source, collector.provider, testLogger(output)); err != nil {
		t.Fatal(err)
	}
	for want := int64(1); want <= 2; want++ {
		got := collector.collect(t)
		if got["harpy.async.observer.failures"] != want {
			t.Fatalf("async observer failures = %d, want %d", got["harpy.async.observer.failures"], want)
		}
	}
	logged := output.String()
	for _, field := range []string{
		`"telemetry.component":"platform_async"`,
		`"telemetry.operation":"observe_backlog"`,
		`"failure.code":"ASYNC_BACKLOG_OBSERVATION_FAILED"`,
	} {
		if !strings.Contains(logged, field) {
			t.Fatalf("async observer log lacks %s: %s", field, logged)
		}
	}
	if strings.Contains(logged, "secret-token") {
		t.Fatalf("async observer log leaked source error: %s", logged)
	}
}

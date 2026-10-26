package telemetry

import (
	"context"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	metricnoop "go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/sdk/resource"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
)

func TestShutdownIsConcurrentAndIdempotent(t *testing.T) {
	var traceCalls atomic.Int32
	var metricCalls atomic.Int32
	var orderMu sync.Mutex
	var order []string
	output := &lockedBuffer{}
	runtime := assembleRuntime(
		tracenoop.NewTracerProvider(), metricnoop.NewMeterProvider(), resource.Empty(),
		newFailureReporter(testLogger(output)),
		func(context.Context) error {
			traceCalls.Add(1)
			orderMu.Lock()
			order = append(order, signalTrace)
			orderMu.Unlock()
			return nil
		},
		func(context.Context) error {
			metricCalls.Add(1)
			orderMu.Lock()
			order = append(order, signalMetric)
			orderMu.Unlock()
			return nil
		},
	)
	var wait sync.WaitGroup
	errors := make(chan error, 32)
	for index := 0; index < cap(errors); index++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			errors <- runtime.Shutdown(context.Background())
		}()
	}
	wait.Wait()
	close(errors)
	for err := range errors {
		if err != nil {
			t.Fatal(err)
		}
	}
	if traceCalls.Load() != 1 || metricCalls.Load() != 1 {
		t.Fatalf("shutdown count: trace=%d metric=%d", traceCalls.Load(), metricCalls.Load())
	}
	if len(order) != 2 || order[0] != signalTrace || order[1] != signalMetric {
		t.Fatalf("shutdown order: %#v", order)
	}
}

func TestShutdownSanitizesAndCachesProviderFailure(t *testing.T) {
	output := &lockedBuffer{}
	runtime := assembleRuntime(
		tracenoop.NewTracerProvider(), metricnoop.NewMeterProvider(), resource.Empty(),
		newFailureReporter(testLogger(output)),
		func(context.Context) error { return secretFailure() },
		func(context.Context) error { return secretFailure() },
	)
	first := runtime.Shutdown(context.Background())
	second := runtime.Shutdown(context.Background())
	if first == nil || second == nil || first.Error() != second.Error() {
		t.Fatalf("shutdown result was not cached: %v / %v", first, second)
	}
	if strings.Contains(first.Error(), "secret-token") || strings.Contains(output.String(), "secret-token") {
		t.Fatalf("provider error leaked: error=%q log=%s", first, output.String())
	}
	if failures := runtime.ExportFailures(); failures.Trace != 1 || failures.Metric != 1 {
		t.Fatalf("unexpected failure counts: %#v", failures)
	}
}

func TestNilShutdownContextDoesNotConsumeLifecycle(t *testing.T) {
	var calls atomic.Int32
	runtime := assembleRuntime(
		tracenoop.NewTracerProvider(), metricnoop.NewMeterProvider(), resource.Empty(),
		newFailureReporter(testLogger(&lockedBuffer{})),
		func(context.Context) error { calls.Add(1); return nil },
		func(context.Context) error { calls.Add(1); return nil },
	)
	if err := runtime.Shutdown(nil); err == nil {
		t.Fatal("nil shutdown context accepted")
	}
	if err := runtime.Shutdown(context.Background()); err != nil {
		t.Fatal(err)
	}
	if calls.Load() != 2 {
		t.Fatalf("valid shutdown did not run: %d", calls.Load())
	}
}

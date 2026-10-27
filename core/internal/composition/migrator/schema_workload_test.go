package migrator

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/admission"
)

type schemaLoopFunc func(context.Context, *admission.Gate) error

func (run schemaLoopFunc) Run(
	ctx context.Context,
	gate *admission.Gate,
) error {
	return run(ctx, gate)
}

func TestSchemaWorkloadFinishesGracefullyAfterDrain(t *testing.T) {
	gate := &admission.Gate{}
	started := make(chan struct{})
	loop := schemaLoopFunc(func(_ context.Context, admission *admission.Gate) error {
		close(started)
		for !admission.Draining() {
			time.Sleep(time.Millisecond)
		}
		return nil
	})
	workload, err := startSchemaWorkload(loop, gate)
	if err != nil {
		t.Fatal(err)
	}
	<-started
	lifecycle := workload.lifecycle()
	lifecycle.BeginDrain()
	if err := lifecycle.Wait(testWaitContext(t)); err != nil {
		t.Fatalf("graceful drain: %v", err)
	}
	if err, open := <-lifecycle.Failures; open || err != nil {
		t.Fatalf("unexpected failure: %v, open=%v", err, open)
	}
}

func TestSchemaWorkloadForwardsUnexpectedFailure(t *testing.T) {
	expected := errors.New("executor failed")
	workload, err := startSchemaWorkload(
		schemaLoopFunc(func(context.Context, *admission.Gate) error {
			return expected
		}),
		&admission.Gate{},
	)
	if err != nil {
		t.Fatal(err)
	}
	if actual := <-workload.lifecycle().Failures; !errors.Is(actual, expected) {
		t.Fatalf("failure=%v", actual)
	}
}

func TestSchemaWorkloadRejectsCleanStopBeforeDrain(t *testing.T) {
	gate := &admission.Gate{}
	gate.Open()
	workload, err := startSchemaWorkload(
		schemaLoopFunc(func(context.Context, *admission.Gate) error {
			return nil
		}),
		gate,
	)
	if err != nil {
		t.Fatal(err)
	}
	if failure := <-workload.lifecycle().Failures; failure == nil {
		t.Fatal("clean runtime stop was not reported as a failure")
	}
}

func TestSchemaWorkloadForceCancelsLoop(t *testing.T) {
	started := make(chan struct{})
	workload, err := startSchemaWorkload(
		schemaLoopFunc(func(ctx context.Context, _ *admission.Gate) error {
			close(started)
			<-ctx.Done()
			return ctx.Err()
		}),
		&admission.Gate{},
	)
	if err != nil {
		t.Fatal(err)
	}
	<-started
	lifecycle := workload.lifecycle()
	if err := lifecycle.Force(); err != nil {
		t.Fatal(err)
	}
	if err := lifecycle.Wait(testWaitContext(t)); err != nil {
		t.Fatalf("forced wait: %v", err)
	}
}

func testWaitContext(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	t.Cleanup(cancel)
	return ctx
}

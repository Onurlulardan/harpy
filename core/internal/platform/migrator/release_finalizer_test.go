package migrator

import (
	"context"
	"errors"
	"testing"
)

type releaseFinalizerFixture struct {
	called bool
	err    error
}

func (fixture *releaseFinalizerFixture) FinalizeRelease(
	context.Context,
	FinalizationQueryRow,
) error {
	fixture.called = true
	return fixture.err
}

func TestFinalizeReleaseWrapsDeterministicFailure(t *testing.T) {
	cause := errors.New("publication conflict")
	fixture := &releaseFinalizerFixture{err: cause}
	runner := &Runner{releaseFinalizer: fixture}
	queryRow := func(context.Context, string, ...any) FinalizationRow {
		t.Fatal("failing fixture unexpectedly queried")
		return nil
	}

	err := runner.finalizeRelease(context.Background(), queryRow)
	if !fixture.called || !errors.Is(err, cause) ||
		!isReleaseFinalizationError(err) {
		t.Fatalf("finalization failure = %v, called=%t", err, fixture.called)
	}
}

func TestFinalizeReleaseAllowsUnconfiguredRunner(t *testing.T) {
	runner := &Runner{}
	if err := runner.finalizeRelease(context.Background(), nil); err != nil {
		t.Fatalf("optional release finalizer = %v", err)
	}
}

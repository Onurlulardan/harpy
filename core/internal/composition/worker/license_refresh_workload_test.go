package worker

import (
	"context"
	"errors"
	"testing"
	"time"
)

type licenseCycleFixture struct {
	err   error
	calls chan struct{}
}

func (fixture *licenseCycleFixture) RunOnce(
	context.Context,
) (time.Duration, error) {
	select {
	case fixture.calls <- struct{}{}:
	default:
	}
	return time.Second, fixture.err
}

func TestLicenseRefreshWorkloadDrainsAndReportsTerminalFailure(t *testing.T) {
	cycle := &licenseCycleFixture{calls: make(chan struct{}, 1)}
	workload, err := newLicenseRefreshWorkload(cycle)
	if err != nil || workload.Start(context.Background()) != nil {
		t.Fatal(err)
	}
	select {
	case <-cycle.calls:
	case <-time.After(time.Second):
		t.Fatal("license refresh cycle did not start")
	}
	workload.BeginDrain()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err = workload.Wait(ctx); err != nil {
		t.Fatal(err)
	}

	terminal := errors.New("license authority fenced")
	failed, err := newLicenseRefreshWorkload(
		&licenseCycleFixture{calls: make(chan struct{}, 1), err: terminal},
	)
	if err != nil || failed.Start(context.Background()) != nil {
		t.Fatal(err)
	}
	select {
	case got := <-failed.Failures():
		if !errors.Is(got, terminal) {
			t.Fatalf("failure=%v", got)
		}
	case <-time.After(time.Second):
		t.Fatal("terminal license failure was not reported")
	}
}

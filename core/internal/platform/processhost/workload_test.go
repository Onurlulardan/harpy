package processhost

import (
	"context"
	"testing"
)

func TestWorkloadRequiresFailureSignal(t *testing.T) {
	workload := Workload{
		BeginDrain: func() {},
		Wait:       func(context.Context) error { return nil },
		Force:      func() error { return nil },
	}
	if err := workload.validate(); err == nil {
		t.Fatal("workload without failure signal accepted")
	}
	workload.Failures = make(chan error)
	if err := workload.validate(); err != nil {
		t.Fatalf("complete workload rejected: %v", err)
	}
}

func TestWorkloadFailureIsForwarded(t *testing.T) {
	source := make(chan error, 1)
	forwarded := make(chan error, 1)
	host := &Host{dependencies: Dependencies{Workload: Workload{
		BeginDrain: func() {}, Wait: func(context.Context) error { return nil },
		Force: func() error { return nil }, Failures: source,
	}}}
	host.startWorkloadMonitor(forwarded)
	expected := context.Canceled
	source <- expected
	if actual := <-forwarded; actual != expected {
		t.Fatalf("forwarded failure=%v", actual)
	}
}

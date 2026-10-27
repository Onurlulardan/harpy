package health

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/admission"
)

func TestPollerUpdatesMonitorAndObserverImmediately(t *testing.T) {
	monitor := newPollerMonitor(t)
	ready := readyHealthStates()
	observed := make(chan Observation, 1)
	poller, err := NewPoller(monitor, time.Second, 100*time.Millisecond,
		func(context.Context) States { return ready },
		func(observation Observation) { observed <- observation },
	)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { defer close(done); poller.Run(ctx) }()
	if got := <-observed; got.States() != ready || !got.Ready() {
		t.Fatalf("observed states = %#v", got.States())
	}
	assertReason(t, monitor, runtimev1.ReadinessReasonNone)
	cancel()
	<-done
}

func TestPollerRepeatsWithoutOverlappingProbes(t *testing.T) {
	monitor := newPollerMonitor(t)
	var calls atomic.Int32
	twoCalls := make(chan struct{})
	poller, err := NewPoller(monitor, 5*time.Millisecond, 5*time.Millisecond,
		func(context.Context) States {
			if calls.Add(1) == 2 {
				close(twoCalls)
			}
			return States{}
		},
		func(Observation) {},
	)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	go poller.Run(ctx)
	select {
	case <-twoCalls:
	case <-time.After(time.Second):
		t.Fatal("periodic probe did not run")
	}
	cancel()
}

func TestPollerFailsClosedWhenProbePanics(t *testing.T) {
	monitor := newPollerMonitor(t)
	observed := make(chan Observation, 1)
	poller, err := NewPoller(monitor, time.Second, time.Second,
		func(context.Context) States { panic("do not expose") },
		func(observation Observation) { observed <- observation },
	)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	go poller.Run(ctx)
	if got := <-observed; got.States() != (States{}) {
		t.Fatalf("panic states = %#v", got.States())
	}
	assertReason(t, monitor, runtimev1.ReadinessReasonDependencyFailed)
	cancel()
}

func TestPollerClosesAdmissionBeforePublishingFailedSnapshot(t *testing.T) {
	now := time.Now()
	monitor, err := NewMonitor(
		runtimev1.ServiceCoreAPI, time.Hour, func() time.Time { return now },
	)
	if err != nil {
		t.Fatal(err)
	}
	ready := readyHealthStates()
	monitor.Update(ready)
	monitor.MarkRunning()
	gate := admission.NewGate(time.Now)
	gate.OpenUntil(now.Add(2 * time.Hour))
	observerEntered := make(chan struct{})
	releaseObserver := make(chan struct{})
	poller, err := NewPoller(
		monitor, time.Hour, time.Second,
		func(context.Context) States { return States{} },
		func(observation Observation) {
			gate.Pause()
			close(observerEntered)
			<-releaseObserver
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { defer close(done); poller.Run(ctx) }()
	<-observerEntered
	if gate.Accepting() {
		t.Fatal("failed observation remained admissible")
	}
	assertReason(t, monitor, runtimev1.ReadinessReasonNone)
	close(releaseObserver)
	cancel()
	<-done
	assertReason(t, monitor, runtimev1.ReadinessReasonDependencyFailed)
}

func TestNewPollerRejectsInvalidInputs(t *testing.T) {
	monitor := newPollerMonitor(t)
	probe := func(context.Context) States { return States{} }
	observe := func(Observation) {}
	if _, err := NewPoller(nil, time.Second, time.Second, probe, observe); err == nil {
		t.Fatal("expected monitor rejection")
	}
	if _, err := NewPoller(monitor, time.Second, 2*time.Second, probe, observe); err == nil {
		t.Fatal("expected timeout rejection")
	}
	if _, err := NewPoller(monitor, time.Second, time.Second, nil, observe); err == nil {
		t.Fatal("expected callback rejection")
	}
}

func newPollerMonitor(t *testing.T) *Monitor {
	t.Helper()
	monitor, err := NewMonitor(runtimev1.ServiceCoreAPI, time.Second, time.Now)
	if err != nil {
		t.Fatal(err)
	}
	return monitor
}

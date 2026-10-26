package lifecycle

import (
	"context"
	"errors"
	"os"
	"reflect"
	"sync"
	"testing"
	"time"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
)

func TestCoordinatorDrainsInContractOrder(t *testing.T) {
	var mu sync.Mutex
	var events []string
	record := func(value string) {
		mu.Lock()
		defer mu.Unlock()
		events = append(events, value)
	}
	hooks := NewHooks()
	registerRecordingHook(t, hooks, record)
	coordinator := mustCoordinator(t, Plan{
		BeginDrain: func() { record("admission") }, InFlight: hooks,
		FlushTelemetry: func(context.Context) error { record("telemetry"); return nil },
		StopManagement: func(context.Context) error { record("management"); return nil },
		ForceClose:     func() error { record("force-close"); return nil }, AbsoluteDeadline: time.Second,
	})
	interrupts := make(chan os.Signal, 2)
	interrupts <- os.Interrupt
	outcome := coordinator.Run(context.Background(), interrupts, make(chan error))
	if outcome.ExitCode() != runtimev1.ExitCodeClean || outcome.Reason() != ReasonSignal {
		t.Fatalf("outcome = %d/%s", outcome.ExitCode(), outcome.Reason())
	}
	mu.Lock()
	defer mu.Unlock()
	want := []string{"admission", "hook-begin", "hook-wait", "telemetry", "management"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestCoordinatorPreservesServeFailureCodeAfterCleanDrain(t *testing.T) {
	coordinator := mustCoordinator(t, noopPlan(time.Second))
	failures := make(chan error, 1)
	failures <- errors.New("listener failed")
	outcome := coordinator.Run(context.Background(), make(chan os.Signal), failures)
	if outcome.ExitCode() != runtimev1.ExitCodeServeFailure || outcome.Reason() != ReasonServeFailure {
		t.Fatalf("outcome = %d/%s", outcome.ExitCode(), outcome.Reason())
	}
}

func TestCoordinatorSecondSignalForcesClose(t *testing.T) {
	waiting := make(chan struct{})
	releaseWait := make(chan struct{})
	forced := make(chan struct{}, 2)
	hooks := NewHooks()
	if err := hooks.Register(Hook{
		Name: "blocked", Begin: func() error { return nil },
		Wait:  func(context.Context) error { close(waiting); <-releaseWait; return nil },
		Force: func() error { close(releaseWait); forced <- struct{}{}; return nil },
	}); err != nil {
		t.Fatalf("register: %v", err)
	}
	plan := noopPlan(time.Second)
	plan.InFlight = hooks
	plan.ForceClose = func() error { forced <- struct{}{}; return nil }
	coordinator := mustCoordinator(t, plan)
	interrupts := make(chan os.Signal, 2)
	interrupts <- os.Interrupt
	result := make(chan Outcome, 1)
	go func() { result <- coordinator.Run(context.Background(), interrupts, make(chan error)) }()
	<-waiting
	interrupts <- os.Interrupt
	outcome := <-result
	if outcome.ExitCode() != runtimev1.ExitCodeDrainForced || outcome.Reason() != ReasonSecondSignal {
		t.Fatalf("outcome = %d/%s", outcome.ExitCode(), outcome.Reason())
	}
	if len(forced) != 2 {
		t.Fatalf("force callbacks = %d", len(forced))
	}
}

func TestCoordinatorDeadlineForcesClose(t *testing.T) {
	plan := noopPlan(15 * time.Millisecond)
	hooks := NewHooks()
	if err := hooks.Register(Hook{
		Name: "blocked", Begin: func() error { return nil },
		Wait:  func(ctx context.Context) error { <-ctx.Done(); return ctx.Err() },
		Force: func() error { return nil },
	}); err != nil {
		t.Fatalf("register: %v", err)
	}
	plan.InFlight = hooks
	coordinator := mustCoordinator(t, plan)
	interrupts := make(chan os.Signal, 1)
	interrupts <- os.Interrupt
	outcome := coordinator.Run(context.Background(), interrupts, make(chan error))
	if outcome.ExitCode() != runtimev1.ExitCodeDrainForced || outcome.Reason() != ReasonDrainDeadline {
		t.Fatalf("outcome = %d/%s", outcome.ExitCode(), outcome.Reason())
	}
}

func noopPlan(deadline time.Duration) Plan {
	return Plan{
		BeginDrain: func() {}, InFlight: NewHooks(),
		FlushTelemetry: func(context.Context) error { return nil },
		StopManagement: func(context.Context) error { return nil },
		ForceClose:     func() error { return nil }, AbsoluteDeadline: deadline,
	}
}

func mustCoordinator(t *testing.T, plan Plan) *Coordinator {
	t.Helper()
	coordinator, err := NewCoordinator(plan)
	if err != nil {
		t.Fatalf("new coordinator: %v", err)
	}
	return coordinator
}

func registerRecordingHook(t *testing.T, hooks *Hooks, record func(string)) {
	t.Helper()
	if err := hooks.Register(Hook{
		Name: "public", Begin: func() error { record("hook-begin"); return nil },
		Wait:  func(context.Context) error { record("hook-wait"); return nil },
		Force: func() error { record("hook-force"); return nil },
	}); err != nil {
		t.Fatalf("register: %v", err)
	}
}

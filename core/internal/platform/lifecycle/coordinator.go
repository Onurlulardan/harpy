package lifecycle

import (
	"context"
	"fmt"
	"os"
	"time"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
)

type Shutdown func(context.Context) error
type ForceClose func() error

type Plan struct {
	BeginDrain       func()
	InFlight         *Hooks
	FlushTelemetry   Shutdown
	StopManagement   Shutdown
	ForceClose       ForceClose
	AbsoluteDeadline time.Duration
}

type Coordinator struct {
	plan Plan
}

func NewCoordinator(plan Plan) (*Coordinator, error) {
	if plan.BeginDrain == nil {
		return nil, fmt.Errorf("begin-drain callback is required")
	}
	if plan.InFlight == nil {
		return nil, fmt.Errorf("in-flight lifecycle hooks are required")
	}
	if plan.FlushTelemetry == nil || plan.StopManagement == nil || plan.ForceClose == nil {
		return nil, fmt.Errorf("lifecycle shutdown callbacks are incomplete")
	}
	if plan.AbsoluteDeadline <= 0 {
		return nil, fmt.Errorf("lifecycle deadline must be positive")
	}
	return &Coordinator{plan: plan}, nil
}

func (coordinator *Coordinator) Run(
	ctx context.Context,
	interrupts <-chan os.Signal,
	serveFailures <-chan error,
) Outcome {
	baseCode, initialReason := awaitTrigger(ctx, interrupts, serveFailures)
	coordinator.plan.BeginDrain()
	drainContext, cancel := context.WithTimeout(context.Background(), coordinator.plan.AbsoluteDeadline)
	defer cancel()

	drained := make(chan error, 1)
	go func() { drained <- coordinator.drain(drainContext) }()

	select {
	case err := <-drained:
		if err != nil {
			if drainContext.Err() != nil {
				coordinator.force()
				return newOutcome(runtimev1.ExitCodeDrainForced, ReasonDrainDeadline)
			}
			coordinator.force()
			return newOutcome(runtimev1.ExitCodeDrainForced, ReasonDrainFailure)
		}
		return newOutcome(baseCode, initialReason)
	case <-interrupts:
		coordinator.force()
		return newOutcome(runtimev1.ExitCodeDrainForced, ReasonSecondSignal)
	case <-drainContext.Done():
		coordinator.force()
		return newOutcome(runtimev1.ExitCodeDrainForced, ReasonDrainDeadline)
	}
}

func (coordinator *Coordinator) drain(ctx context.Context) error {
	if err := coordinator.plan.InFlight.Begin(); err != nil {
		return err
	}
	if err := coordinator.plan.InFlight.Wait(ctx); err != nil {
		return err
	}
	if err := coordinator.plan.FlushTelemetry(ctx); err != nil {
		return err
	}
	return coordinator.plan.StopManagement(ctx)
}

func (coordinator *Coordinator) force() {
	_ = coordinator.plan.InFlight.Force()
	_ = coordinator.plan.ForceClose()
}

func awaitTrigger(
	ctx context.Context,
	interrupts <-chan os.Signal,
	serveFailures <-chan error,
) (runtimev1.ExitCode, Reason) {
	select {
	case <-ctx.Done():
		return runtimev1.ExitCodeClean, ReasonContext
	case <-interrupts:
		return runtimev1.ExitCodeClean, ReasonSignal
	case <-serveFailures:
		return runtimev1.ExitCodeServeFailure, ReasonServeFailure
	}
}

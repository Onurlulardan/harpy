package adoption

import (
	"context"
	"reflect"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/usage/domain"
)

func (coordinator *Coordinator) checkSource(ctx context.Context, run Run,
	request ExecuteRequest, manifest Manifest,
) error {
	current, err := coordinator.source.Inspect(ctx)
	if err != nil || current.Validate() != nil {
		return domain.ErrUnavailable
	}
	if current.Same(manifest) {
		return nil
	}
	_, claimErr := coordinator.target.Claim(ctx, ClaimCommand{OperationID: request.OperationID,
		RunID: request.RunID, LeaseOwner: request.LeaseOwner, ExpectedRevision: run.Revision,
		Manifest: current, ObservedAt: coordinator.clock.Now().UTC()})
	if claimErr != nil {
		return claimErr
	}
	return domain.ErrSourceDrift
}

func (coordinator *Coordinator) finish(ctx context.Context, run Run,
	request ExecuteRequest, manifest Manifest,
) (Run, error) {
	if err := coordinator.checkSource(ctx, run, request, manifest); err != nil {
		return run, err
	}
	command := leaseCommand(run, request, manifest, coordinator.clock.Now().UTC())
	reconciled, err := coordinator.target.Reconcile(ctx, command)
	if err != nil || reconciled.Validate() != nil {
		return reconciled, errOrUnavailable(err)
	}
	command.LeaseGeneration = reconciled.LeaseGeneration
	command.ObservedAt = coordinator.clock.Now().UTC()
	completed, err := coordinator.target.Complete(ctx, command)
	if err != nil || completed.Validate() != nil || completed.Status != StatusCompleted {
		return completed, errOrUnavailable(err)
	}
	return completed, nil
}

func leaseCommand(run Run, request ExecuteRequest, manifest Manifest,
	observedAt time.Time,
) LeaseCommand {
	return LeaseCommand{OperationID: request.OperationID, RunID: request.RunID,
		LeaseOwner: request.LeaseOwner, LeaseGeneration: run.LeaseGeneration,
		Manifest: manifest, ObservedAt: observedAt}
}

func validateRun(run Run, request ExecuteRequest, manifest Manifest) error {
	if run.Validate() != nil || run.OperationID != request.OperationID ||
		run.RunID != request.RunID || !run.Manifest.Same(manifest) ||
		run.TransformSHA256 != TransformSHA256 {
		return domain.ErrUnavailable
	}
	if run.Status == StatusFenced {
		return domain.ErrSourceDrift
	}
	return nil
}

func errOrUnavailable(err error) error {
	if err != nil {
		return err
	}
	return domain.ErrUnavailable
}

func nilPort(value any) bool {
	if value == nil {
		return true
	}
	seen := reflect.ValueOf(value)
	switch seen.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return seen.IsNil()
	default:
		return false
	}
}

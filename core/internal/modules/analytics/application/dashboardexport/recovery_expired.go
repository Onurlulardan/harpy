package dashboardexport

import (
	"context"
	"errors"
)

func (runner *RecoveryRunner) recoverExpired(
	scan, force context.Context,
) (bool, error) {
	candidate, found, err := runner.store.NextExpired(scan)
	if err != nil || !found {
		return false, err
	}
	if scan.Err() != nil {
		return false, nil
	}
	result, err := runner.store.FinalizeExpired(force, candidate, runner.clock.Now())
	if errors.Is(err, ErrDispatchCandidateChanged) ||
		errors.Is(err, ErrDispatchLeaseStale) {
		return true, nil
	}
	if err != nil {
		if force.Err() != nil {
			return true, nil
		}
		return true, err
	}
	if result.State == RenderExecutionOutcomeUnknown {
		return true, nil
	}
	switch result.State {
	case RenderExecutionRequeued, RenderExecutionDead, RenderExecutionCancelled:
		return true, runner.reapResolved(force)
	default:
		return true, ErrUnavailable
	}
}

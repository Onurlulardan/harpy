package schemaexecutor

import (
	"context"
	"errors"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/platform/admission"
)

func (executor *Executor) processOne(
	ctx context.Context,
	permit admission.Permit,
) (bool, error) {
	if !permit.Granted() {
		return false, fmt.Errorf("schema executor admission permit is required")
	}
	recovered, err := executor.persistence.RecoverExpired(ctx)
	if err != nil {
		return false, err
	}
	leaseID, err := executor.newLeaseID()
	if err != nil {
		return false, err
	}
	record, found, err := executor.persistence.Claim(ctx, ClaimRequest{
		ExecutorID: executor.executorID, LeaseID: leaseID,
	})
	if err != nil || !found {
		return recovered, err
	}
	claim, err := acceptClaim(record)
	if err != nil {
		return true, err
	}
	if err := executor.persistence.MarkRunning(ctx, claim.lease); err != nil {
		return true, err
	}
	applyContext, cancelApply := context.WithCancel(ctx)
	heartbeatDone := make(chan error, 1)
	go executor.heartbeat(applyContext, cancelApply, claim.lease, heartbeatDone)
	plan, validationErr := ValidateClaim(claim)
	if validationErr != nil {
		cancelApply()
		<-heartbeatDone
		return true, executor.recordFailure(ctx, claim.lease, validationErr)
	}
	applyErr := executor.persistence.Apply(applyContext, ApplyRequest{
		Lease: claim.lease, Plan: plan,
	})
	cancelApply()
	heartbeatErr := <-heartbeatDone
	if applyErr == nil {
		return true, nil
	}
	if errors.Is(ctx.Err(), context.Canceled) {
		return true, ctx.Err()
	}
	if heartbeatErr != nil {
		return true, heartbeatErr
	}
	kind, _ := FaultDetails(applyErr)
	if kind == FaultLeaseLost {
		return true, applyErr
	}
	return true, executor.recordFailure(ctx, claim.lease, applyErr)
}

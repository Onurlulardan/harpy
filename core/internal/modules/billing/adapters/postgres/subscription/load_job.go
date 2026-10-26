package subscription

import (
	"context"
	"fmt"

	subscriptionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/subscription"
	subscriptiondomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/subscription"
	"github.com/google/uuid"
)

func (storage *store) loadJob(
	ctx context.Context,
	authorityID uuid.UUID,
	execution subscriptionapp.JobExecution,
) (subscriptionapp.LockedState, *receiptRecord, error) {
	locked, receipt, err := storage.loadBase(ctx, authorityID, execution.Execution)
	if err != nil {
		return subscriptionapp.LockedState{}, nil, err
	}
	if receipt != nil && !receipt.matches(execution.Execution, string(modeJob)) {
		return subscriptionapp.LockedState{}, nil, subscriptionapp.ErrIdempotencyReused
	}
	job, err := storage.assertJob(ctx, authorityID, execution)
	if err != nil {
		return subscriptionapp.LockedState{}, nil, err
	}
	if receipt != nil {
		return locked, receipt, nil
	}
	jobs, err := storage.listBoundaryJobs(ctx, authorityID, execution.SubscriptionID)
	if err != nil {
		return subscriptionapp.LockedState{}, nil, err
	}
	if !containsAssertedJob(jobs, execution, job) {
		return subscriptionapp.LockedState{}, nil, subscriptionapp.ErrBoundaryMismatch
	}
	locked.BoundaryJobs = jobs
	boundary, pending, err := storage.loadBoundary(
		ctx, authorityID, execution, job, locked.Subscription.State,
	)
	if err != nil {
		return subscriptionapp.LockedState{}, nil, err
	}
	locked.Boundary, locked.PendingChange = boundary, pending
	return locked, nil, nil
}

func containsAssertedJob(
	jobs []subscriptionapp.JobIntent,
	execution subscriptionapp.JobExecution,
	asserted assertedJob,
) bool {
	for _, job := range jobs {
		if job.ID == execution.JobID && job.EffectID == asserted.effectID &&
			job.SubscriptionID == asserted.subscriptionID &&
			job.Boundary == asserted.boundary && job.RunAt.Equal(asserted.runAt) &&
			job.ExpectedRevision == asserted.expectedRevision &&
			job.PayloadRevision == asserted.payloadRevision {
			return true
		}
	}
	return false
}

func (storage *store) loadBoundary(
	ctx context.Context,
	authorityID uuid.UUID,
	execution subscriptionapp.JobExecution,
	job assertedJob,
	state subscriptiondomain.State,
) (*subscriptionapp.BoundarySnapshot, *subscriptionapp.SubscriptionChange, error) {
	boundary := &subscriptionapp.BoundarySnapshot{
		Kind: execution.Boundary, DueAt: job.runAt,
	}
	switch execution.Boundary {
	case subscriptionapp.BoundaryCancellation, subscriptionapp.BoundaryGraceEnd:
		return boundary, nil, nil
	case subscriptionapp.BoundaryManualChange:
		return storage.loadChangeBoundary(ctx, authorityID, execution, job, state, boundary)
	case subscriptionapp.BoundaryTrialEnd, subscriptionapp.BoundaryPeriodEnd:
		evidence, err := storage.findBoundaryEvidence(ctx, authorityID, execution, job)
		if err != nil {
			return nil, nil, err
		}
		if err := evidence.apply(boundary); err != nil {
			return nil, nil, err
		}
		return boundary, nil, nil
	default:
		return nil, nil, fmt.Errorf("unsupported billing boundary %q", execution.Boundary)
	}
}

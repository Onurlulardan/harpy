package subscriptionapp

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type ExecuteDueBoundaryRequest struct {
	SubscriptionID   uuid.UUID
	JobID            uuid.UUID
	EffectID         uuid.UUID
	ExpectedRevision int64
	LeaseGeneration  int64
	Boundary         BoundaryKind
	DueAt            time.Time
}

func (commands *Commands) ExecuteDueBoundary(
	ctx context.Context,
	request ExecuteDueBoundaryRequest,
) (Result, error) {
	now, err := commands.instant()
	if err != nil || ctx == nil || ctx.Err() != nil || !validID(request.SubscriptionID) ||
		!validV7(request.JobID) || !validV7(request.EffectID) ||
		request.ExpectedRevision < 1 || request.LeaseGeneration < 1 ||
		!request.Boundary.valid() || !validInstant(request.DueAt) {
		return Result{}, ErrRequest
	}
	digest := dueDigest(request)
	execution := JobExecution{
		Execution: Execution{
			OperationID: OperationExecuteDue, SubscriptionID: request.SubscriptionID,
			SourceOperationID: request.EffectID,
			ExpectedRevision:  request.ExpectedRevision, RequestSHA256: digest,
		},
		JobID: request.JobID, LeaseGeneration: request.LeaseGeneration,
		Boundary: request.Boundary,
	}
	result, err := commands.store.Job(ctx, execution, func(locked LockedState) (CommitPlan, error) {
		if !locked.valid() || locked.Subscription.ID != request.SubscriptionID {
			return CommitPlan{}, ErrSnapshot
		}
		if replay, found, replayErr := receiptReplay(locked.Receipt, execution.Execution); found {
			return replay, replayErr
		}
		if locked.Boundary == nil {
			return CommitPlan{}, ErrSnapshot
		}
		if locked.Subscription.Revision != request.ExpectedRevision {
			return CommitPlan{}, ErrStaleRevision
		}
		boundary := *locked.Boundary
		if boundary.Kind != request.Boundary || !boundary.DueAt.Equal(request.DueAt) {
			return CommitPlan{}, ErrBoundaryMismatch
		}
		if now.Before(request.DueAt) {
			return CommitPlan{}, ErrBoundaryNotDue
		}
		after, resolution, transitionErr := dueSnapshot(locked.Subscription, boundary, now)
		if transitionErr != nil {
			return CommitPlan{}, transitionErr
		}
		identifiers, identifierErr := commands.baseIdentifiers()
		if identifierErr != nil {
			return CommitPlan{}, identifierErr
		}
		plan := basePlan(planInputs{
			Before: locked.Subscription, After: after, Execution: execution.Execution,
			ReceiptID: identifiers.receipt, AuditID: identifiers.audit,
			EventID:            identifiers.subscriptionEvent,
			EntitlementEventID: identifiers.entitlementEvent, Now: now,
			EventActorID: execution.SourceOperationID,
			Kind:         "due_boundary_applied", Reason: string(boundary.Kind),
		})
		plan.CompleteJob = &JobCompletion{
			ID: request.JobID, LeaseGeneration: request.LeaseGeneration, CompletedAt: now,
		}
		plan.CancelJobs = boundaryCancellationsExcept(
			locked, request.JobID, now, "due_boundary_superseded",
		)
		plan.ResolveChange = resolution
		if kind, runAt, scheduled := nextBoundary(after); scheduled {
			job, jobErr := commands.boundaryJob(after.ID, kind, runAt, after.Revision)
			if jobErr != nil {
				return CommitPlan{}, jobErr
			}
			plan.Jobs = []JobIntent{job}
		}
		if !plan.valid() {
			return CommitPlan{}, ErrConfiguration
		}
		return plan, nil
	})
	if err != nil {
		return Result{}, fmt.Errorf("execute billing subscription due boundary: %w", err)
	}
	return checkedResult(result, request.SubscriptionID)
}

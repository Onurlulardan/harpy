package subscriptionapp

import (
	"context"
	"errors"
	"sync"

	"github.com/google/uuid"
)

var errFakeCommit = errors.New("fake atomic commit failed")

type fakeStore struct {
	mu            sync.Mutex
	snapshot      SubscriptionSnapshot
	receipts      map[uuid.UUID]OperationReceipt
	target        *TargetPrice
	boundaryJobs  []JobIntent
	pendingChange *SubscriptionChange
	trialClaim    *TrialClaim
	boundary      *BoundarySnapshot
	plans         []CommitPlan
	outbox        []OutboxIntent
	audits        []AuditIntent
	failCommit    bool
}

func (store *fakeStore) Actor(
	ctx context.Context,
	execution ActorExecution,
	planner Planner,
) (Result, error) {
	if !actorExecutionValid(execution) {
		return Result{}, ErrRequest
	}
	return store.atomic(ctx, execution.Execution, planner)
}

func (store *fakeStore) Provider(
	ctx context.Context,
	execution ProviderExecution,
	planner Planner,
) (Result, error) {
	if !execution.valid() {
		return Result{}, ErrRequest
	}
	return store.atomic(ctx, execution.Execution, planner)
}

func (store *fakeStore) Job(
	ctx context.Context,
	execution JobExecution,
	planner Planner,
) (Result, error) {
	if !execution.valid() {
		return Result{}, ErrRequest
	}
	return store.atomic(ctx, execution.Execution, planner)
}

func actorExecutionValid(execution ActorExecution) bool {
	switch execution.OperationID {
	case OperationCancelAtPeriodEnd, OperationChangeAtPeriodEnd,
		OperationClaimTrial, OperationPlatformTransition,
		OperationResumeCancellation:
		return execution.valid(execution.OperationID)
	default:
		return false
	}
}

func (store *fakeStore) atomic(
	ctx context.Context,
	execution Execution,
	planner Planner,
) (Result, error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	if ctx == nil || ctx.Err() != nil || planner == nil {
		return Result{}, ErrRequest
	}
	locked := store.locked(execution.SourceOperationID)
	plan, err := planner(locked)
	if err != nil {
		return Result{}, err
	}
	if !plan.valid() {
		return Result{}, ErrConfiguration
	}
	if plan.Replay {
		return Result{Subscription: plan.After, Replay: true}, nil
	}
	if store.failCommit {
		return Result{}, errFakeCommit
	}
	if plan.BeforeRevision != store.snapshot.Revision ||
		plan.After.ID != store.snapshot.ID || plan.After.TenantID != store.snapshot.TenantID {
		return Result{}, ErrStaleRevision
	}
	store.snapshot = plan.After
	store.receipts[plan.Receipt.SourceOperationID] = plan.Receipt
	store.applyJobs(plan)
	if plan.Change != nil {
		copy := *plan.Change
		store.pendingChange = &copy
	}
	if plan.ResolveChange != nil {
		store.pendingChange = nil
	}
	if plan.TrialClaim != nil {
		copy := *plan.TrialClaim
		store.trialClaim = &copy
	}
	store.plans = append(store.plans, plan)
	store.outbox = append(store.outbox, plan.Outbox...)
	store.audits = append(store.audits, plan.Audit)
	return Result{Subscription: plan.After}, nil
}

func (store *fakeStore) locked(operationID uuid.UUID) LockedState {
	locked := LockedState{
		Subscription: store.snapshot, TargetPrice: store.target,
		BoundaryJobs:  append([]JobIntent(nil), store.boundaryJobs...),
		PendingChange: store.pendingChange, TrialClaim: store.trialClaim,
		Boundary: store.boundary,
	}
	if receipt, found := store.receipts[operationID]; found {
		copy := receipt
		locked.Receipt = &copy
	}
	return locked
}

func (store *fakeStore) applyJobs(plan CommitPlan) {
	for _, cancellation := range plan.CancelJobs {
		store.removeJob(cancellation.ID)
	}
	if plan.CompleteJob != nil {
		store.removeJob(plan.CompleteJob.ID)
	}
	store.boundaryJobs = append(store.boundaryJobs, plan.Jobs...)
}

func (store *fakeStore) removeJob(identifier uuid.UUID) {
	for index := range store.boundaryJobs {
		if store.boundaryJobs[index].ID == identifier {
			store.boundaryJobs = append(store.boundaryJobs[:index], store.boundaryJobs[index+1:]...)
			return
		}
	}
}

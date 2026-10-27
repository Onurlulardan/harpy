package adoption

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Clock interface{ Now() time.Time }

type Worker struct {
	source Source
	target Target
	clock  Clock
}

type StartRequest struct {
	OperationID, RunID uuid.UUID
	Locator            VerifiedLocator
}

type ProcessRequest struct {
	OperationID, RunID, LeaseOwner uuid.UUID
	ExpectedRevision               int64
}

func NewWorker(source Source, target Target, clock Clock) (*Worker, error) {
	if nilPort(source) || nilPort(target) || nilPort(clock) {
		return nil, ErrUnavailable
	}
	return &Worker{source: source, target: target, clock: clock}, nil
}

func (worker *Worker) Execute(ctx context.Context, request ProcessRequest) (Run, error) {
	if worker == nil || ctx == nil || ctx.Err() != nil || request.OperationID == uuid.Nil ||
		request.RunID == uuid.Nil || request.LeaseOwner == uuid.Nil || request.ExpectedRevision < 1 {
		return Run{}, ErrInvalid
	}
	now := worker.clock.Now().UTC()
	if !validObservedAt(now) {
		return Run{}, ErrUnavailable
	}
	run, err := worker.target.AdoptionClaim(ctx, ClaimCommand{OperationID: request.OperationID,
		RunID: request.RunID, LeaseOwner: request.LeaseOwner,
		ExpectedRevision: request.ExpectedRevision, ObservedAt: now})
	if err != nil || !run.valid() {
		return Run{}, normalizeTargetResult(err)
	}
	if run.Status == RunCompleted {
		return run, nil
	}
	if run.Status == RunReconciling {
		completed, completeErr := worker.target.AdoptionComplete(ctx, LeaseCommand{
			OperationID: run.OperationID, RunID: run.RunID, LeaseOwner: run.LeaseOwner,
			LeaseGeneration: run.LeaseGeneration, ObservedAt: now})
		if completeErr != nil || !completed.valid() || completed.Status != RunCompleted {
			return Run{}, normalizeTargetResult(completeErr)
		}
		return completed, nil
	}
	locator, err := run.locator()
	if err != nil {
		return Run{}, ErrUnavailable
	}
	run, done, err := worker.copyPage(ctx, locator, run, now)
	if err != nil || !done {
		return run, err
	}
	manifest, err := worker.source.Inspect(ctx, locator)
	expected := SourceManifest{SnapshotID: run.SnapshotID, RowCount: run.ExpectedCount,
		RootSHA256: run.ExpectedSourceSHA256, ManifestSHA256: run.ManifestSHA256}
	if err != nil || !manifest.valid() || !exactManifest(manifest, expected) ||
		run.CopiedCount != manifest.RowCount || run.SourceSHA256 != manifest.RootSHA256 {
		return Run{}, ErrSourceDrift
	}
	lease := LeaseCommand{OperationID: run.OperationID, RunID: run.RunID,
		LeaseOwner: request.LeaseOwner, LeaseGeneration: run.LeaseGeneration, ObservedAt: now}
	if run, err = worker.target.AdoptionReconcile(ctx, lease); err != nil {
		return Run{}, err
	}
	if !run.valid() || run.Status != RunReconciling {
		return Run{}, ErrUnavailable
	}
	run, err = worker.target.AdoptionComplete(ctx, lease)
	if err != nil || !run.valid() || run.Status != RunCompleted {
		return Run{}, normalizeTargetResult(err)
	}
	return run, nil
}

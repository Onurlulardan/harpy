package customerlifecycle

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
)

func (worker *Worker) RunOnce(ctx context.Context, workerID uuid.UUID) (Operation, error) {
	if worker == nil || ctx == nil || ctx.Err() != nil || workerID == uuid.Nil {
		return Operation{}, ErrInvalid
	}
	now := worker.dependencies.Clock.Now().UTC().Truncate(microsecond)
	claim, err := worker.dependencies.Store.Claim(ctx, workerID, now, 2*time.Minute)
	if err != nil {
		return Operation{}, err
	}
	switch claim.Kind {
	case OperationOffboard:
		return worker.dependencies.Store.CompleteOffboard(ctx, claim, now)
	case OperationPurge:
		if !claim.Proof.Valid(now) || strings.TrimSpace(claim.SchemaName) == "" {
			_ = worker.dependencies.Store.Fail(ctx, claim, "PURGE_PROOF_INVALID", now)
			return Operation{}, ErrPurgeProof
		}
		receipt, executeErr := worker.dependencies.Executor.Purge(ctx, claim)
		completedAt := worker.dependencies.Clock.Now().UTC().Truncate(microsecond)
		if executeErr != nil || completedAt.Before(now) || !receipt.Valid(completedAt) {
			_ = worker.dependencies.Store.Fail(ctx, claim, "PURGE_EXECUTION_FAILED", now)
			return Operation{}, ErrUnavailable
		}
		return worker.dependencies.Store.CompletePurge(ctx, claim, receipt, completedAt)
	default:
		_ = worker.dependencies.Store.Fail(ctx, claim, "OPERATION_KIND_INVALID", now)
		return Operation{}, ErrUnavailable
	}
}

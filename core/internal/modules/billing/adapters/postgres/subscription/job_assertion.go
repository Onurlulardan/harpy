package subscription

import (
	"context"
	"fmt"
	"time"

	subscriptionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/subscription"
	"github.com/google/uuid"
)

type assertedJob struct {
	effectID         uuid.UUID
	subscriptionID   uuid.UUID
	boundary         subscriptionapp.BoundaryKind
	runAt            time.Time
	expectedRevision int64
	payloadRevision  int
}

func (storage *store) assertJob(
	ctx context.Context,
	authorityID uuid.UUID,
	execution subscriptionapp.JobExecution,
) (assertedJob, error) {
	var result assertedJob
	var boundary string
	err := storage.queryRow(
		ctx, assertJobSQL, authorityID, storage.tenantID, execution.JobID,
		execution.SubscriptionID, execution.SourceOperationID,
		execution.ExpectedRevision, string(execution.Boundary),
		execution.LeaseGeneration,
	).Scan(
		&result.effectID, &result.subscriptionID, &boundary, &result.runAt,
		&result.expectedRevision, &result.payloadRevision,
	)
	if err != nil {
		return assertedJob{}, fmt.Errorf("assert billing subscription job: %w", err)
	}
	result.boundary = subscriptionapp.BoundaryKind(boundary)
	result.runAt = result.runAt.UTC()
	if result.effectID != execution.SourceOperationID ||
		result.subscriptionID != execution.SubscriptionID ||
		result.boundary != execution.Boundary ||
		result.expectedRevision != execution.ExpectedRevision ||
		result.payloadRevision != 1 {
		return assertedJob{}, subscriptionapp.ErrBoundaryMismatch
	}
	return result, nil
}

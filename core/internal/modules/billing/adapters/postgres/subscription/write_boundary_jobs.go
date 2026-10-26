package subscription

import (
	"context"
	"fmt"

	subscriptionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/subscription"
	"github.com/google/uuid"
)

func (storage *store) writeBoundaryJobs(
	ctx context.Context,
	authorityID uuid.UUID,
	plan subscriptionapp.CommitPlan,
) error {
	for _, job := range plan.Jobs {
		artifact, err := storage.effects.EnqueueBoundary(ctx, job)
		if err != nil {
			return fmt.Errorf("enqueue billing subscription boundary: %w", err)
		}
		if !artifact.valid() {
			return subscriptionapp.ErrSnapshot
		}
		err = storage.expectID(
			ctx, insertBoundaryJobSQL, job.ID,
			authorityID, storage.tenantID, job.ID, job.SubscriptionID,
			job.EffectID, string(job.Boundary), job.RunAt,
			job.ExpectedRevision, job.PayloadRevision,
			artifact.PayloadReference, artifact.PayloadSHA256[:],
			plan.Audit.OccurredAt,
		)
		if err != nil {
			return fmt.Errorf("insert billing subscription boundary intent: %w", err)
		}
	}
	return nil
}

func (storage *store) cancelBoundaryJobs(
	ctx context.Context,
	plan subscriptionapp.CommitPlan,
) error {
	for _, cancellation := range plan.CancelJobs {
		if err := storage.effects.CancelBoundary(ctx, cancellation); err != nil {
			return fmt.Errorf("cancel billing subscription boundary: %w", err)
		}
	}
	return nil
}

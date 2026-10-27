package subscription

import (
	"context"

	subscriptionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/subscription"
	"github.com/google/uuid"
)

func (storage *store) commit(
	ctx context.Context,
	authorityID uuid.UUID,
	mode bindingMode,
	plan subscriptionapp.CommitPlan,
) error {
	if err := storage.resolveChange(
		ctx, authorityID, plan.After.ID, plan.ResolveChange,
	); err != nil {
		return err
	}
	if err := storage.writeSubscription(ctx, authorityID, plan); err != nil {
		return err
	}
	if err := storage.writeChange(ctx, authorityID, plan.Change); err != nil {
		return err
	}
	if err := storage.writeTrialClaim(ctx, authorityID, plan.TrialClaim); err != nil {
		return err
	}
	if err := storage.writeReceipt(ctx, authorityID, string(mode), plan); err != nil {
		return err
	}
	if err := storage.writeAudit(ctx, authorityID, string(mode), plan); err != nil {
		return err
	}
	if err := storage.writeBoundaryJobs(ctx, authorityID, plan); err != nil {
		return err
	}
	if err := storage.cancelBoundaryJobs(ctx, plan); err != nil {
		return err
	}
	return storage.writeOutbox(ctx, plan)
}

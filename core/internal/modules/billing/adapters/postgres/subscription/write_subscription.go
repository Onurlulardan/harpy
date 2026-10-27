package subscription

import (
	"context"
	"fmt"

	subscriptionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/subscription"
	"github.com/google/uuid"
)

func (storage *store) writeSubscription(
	ctx context.Context,
	authorityID uuid.UUID,
	plan subscriptionapp.CommitPlan,
) error {
	var periodStart, periodEnd any
	if plan.After.Period != nil {
		periodStart = plan.After.Period.Start()
		periodEnd = plan.After.Period.End()
	}
	err := storage.expectID(
		ctx, updateSubscriptionSQL, plan.After.ID,
		authorityID, storage.tenantID, plan.After.ID, plan.BeforeRevision,
		plan.After.OfferingVersionID, plan.After.PriceVersionID,
		string(plan.After.State), periodStart, periodEnd,
		optionalInstantValue(plan.After.TrialEndsAt),
		optionalInstantValue(plan.After.GraceEndsAt),
		plan.After.CancelAtPeriodEnd, optionalInstantValue(plan.After.CancelAt),
		plan.After.UpdatedAt, plan.After.Revision,
	)
	if err != nil {
		return fmt.Errorf("update billing subscription: %w", err)
	}
	return nil
}

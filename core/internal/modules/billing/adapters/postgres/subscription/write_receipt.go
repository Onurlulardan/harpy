package subscription

import (
	"context"
	"fmt"

	subscriptionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/subscription"
	"github.com/google/uuid"
)

func (storage *store) writeReceipt(
	ctx context.Context,
	authorityID uuid.UUID,
	sourceKind string,
	plan subscriptionapp.CommitPlan,
) error {
	var periodStart, periodEnd any
	if plan.After.Period != nil {
		periodStart, periodEnd = plan.After.Period.Start(), plan.After.Period.End()
	}
	receipt := plan.Receipt
	err := storage.expectID(
		ctx, insertReceiptSQL, receipt.ID,
		authorityID, storage.tenantID, receipt.ID, plan.After.ID,
		receipt.SourceOperationID, sourceKind, receipt.OperationID,
		receipt.RequestSHA256[:], receipt.ResultSHA256[:], plan.BeforeRevision,
		plan.After.Revision, string(plan.After.State), receipt.CompletedAt,
		plan.After.ProviderAccountID, plan.After.OfferingVersionID,
		plan.After.PriceVersionID, plan.After.Currency, periodStart, periodEnd,
		optionalInstantValue(plan.After.TrialEndsAt),
		optionalInstantValue(plan.After.GraceEndsAt),
		plan.After.CancelAtPeriodEnd, optionalInstantValue(plan.After.CancelAt),
		plan.After.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert billing subscription receipt: %w", err)
	}
	return nil
}

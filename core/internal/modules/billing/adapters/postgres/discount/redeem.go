package discount

import (
	"context"
	"fmt"
	"time"

	discountapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/discount"
)

func (storage *store) Redeem(
	ctx context.Context,
	execution discountapp.RedeemExecution,
	planner discountapp.RedeemPlanner,
) (discountapp.RedeemResult, error) {
	if !storage.usable(ctx) || planner == nil || execution.Now.IsZero() ||
		execution.Now.Location() != time.UTC || execution.Now.Nanosecond()%1_000 != 0 {
		return discountapp.RedeemResult{}, discountapp.ErrConfiguration
	}
	request := execution.Request
	snapshot, err := storage.lockEligibility(ctx, request.Code, request.SubscriptionID,
		request.OfferingVersionID, request.PriceVersionID, execution.Now)
	if err != nil {
		return discountapp.RedeemResult{}, fmt.Errorf("lock eligibility: %w", err)
	}
	existing, found, err := storage.findRedemption(ctx, request.SourceOperationID)
	if err != nil {
		return discountapp.RedeemResult{}, fmt.Errorf("find redemption: %w", err)
	}
	var existingPointer *discountapp.Redemption
	if found {
		existingPointer = &existing
	}
	plan, err := planner(snapshot, existingPointer)
	if err != nil {
		return discountapp.RedeemResult{}, err
	}
	if plan.Create == nil {
		if !plan.Result.Replay || existingPointer == nil {
			return discountapp.RedeemResult{}, discountapp.ErrSnapshot
		}
		return plan.Result, nil
	}
	if plan.Result.Replay || plan.Result.Redemption != *plan.Create {
		return discountapp.RedeemResult{}, discountapp.ErrSnapshot
	}
	if err := storage.createRedemption(ctx, *plan.Create); err != nil {
		return discountapp.RedeemResult{}, fmt.Errorf("create redemption: %w", err)
	}
	return plan.Result, nil
}

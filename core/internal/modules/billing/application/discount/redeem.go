package discount

import (
	"context"
	"errors"
	"fmt"
	"time"
)

func (commands *Commands) Redeem(
	ctx context.Context,
	request RedeemRequest,
) (RedeemResult, error) {
	now := commands.clock()
	if !validInstant(now) || !request.Code.Valid() || !validV7(request.SubscriptionID) ||
		!validV7(request.OfferingVersionID) || !validV7(request.PriceVersionID) ||
		!validV7(request.SourceOperationID) || !currencyPattern.MatchString(request.Currency) {
		return RedeemResult{}, ErrRequest
	}
	result, err := commands.store.Redeem(
		ctx, RedeemExecution{Request: request, Now: now},
		func(snapshot EligibilitySnapshot, existing *Redemption) (RedemptionPlan, error) {
			return commands.planRedemption(request, now, snapshot, existing)
		},
	)
	if err != nil {
		if errors.Is(err, ErrSnapshot) || errors.Is(err, ErrIneligible) ||
			errors.Is(err, ErrOperationReused) || errors.Is(err, ErrConfiguration) {
			return RedeemResult{}, err
		}
		return RedeemResult{}, fmt.Errorf("redeem billing discount: %w", err)
	}
	return result, nil
}

func (commands *Commands) planRedemption(
	request RedeemRequest,
	now time.Time,
	snapshot EligibilitySnapshot,
	existing *Redemption,
) (RedemptionPlan, error) {
	if !snapshot.validFor(request, now) {
		return RedemptionPlan{}, ErrSnapshot
	}
	if existing != nil {
		result, err := redemptionReplay(*existing, request, snapshot)
		return RedemptionPlan{Result: result}, err
	}
	if !snapshot.OfferingAllowed || exhausted(snapshot.GlobalUseLimit, snapshot.GlobalUsed) ||
		exhausted(snapshot.TenantUseLimit, snapshot.TenantUsed) {
		return RedemptionPlan{}, ErrIneligible
	}
	identifier, err := commands.nextID()
	if err != nil || !validV7(identifier) {
		return RedemptionPlan{}, ErrConfiguration
	}
	redemption := Redemption{
		ID: identifier, DiscountVersionID: snapshot.DiscountVersionID,
		SubscriptionID:    request.SubscriptionID,
		OfferingVersionID: request.OfferingVersionID, PriceVersionID: request.PriceVersionID,
		SourceOperationID: request.SourceOperationID,
		PolicySHA256:      snapshot.PolicySHA256, RedeemedAt: now,
	}
	return RedemptionPlan{
		Result: RedeemResult{Redemption: redemption}, Create: &redemption,
	}, nil
}

func redemptionReplay(
	redemption Redemption,
	request RedeemRequest,
	snapshot EligibilitySnapshot,
) (RedeemResult, error) {
	if !validV7(redemption.ID) || redemption.DiscountVersionID != snapshot.DiscountVersionID ||
		redemption.SubscriptionID != request.SubscriptionID ||
		redemption.OfferingVersionID != request.OfferingVersionID ||
		redemption.PriceVersionID != request.PriceVersionID ||
		redemption.SourceOperationID != request.SourceOperationID ||
		redemption.PolicySHA256 != snapshot.PolicySHA256 || !validInstant(redemption.RedeemedAt) {
		return RedeemResult{}, ErrOperationReused
	}
	return RedeemResult{Redemption: redemption, Replay: true}, nil
}

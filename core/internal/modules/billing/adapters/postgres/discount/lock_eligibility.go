package discount

import (
	"context"
	"fmt"
	"time"

	discountapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/discount"
	discountdomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/discount"
	"github.com/google/uuid"
)

func (storage *store) lockEligibility(
	ctx context.Context,
	code discountdomain.Code,
	subscriptionID uuid.UUID,
	offeringVersionID uuid.UUID,
	priceVersionID uuid.UUID,
	now time.Time,
) (discountapp.EligibilitySnapshot, error) {
	authorityID, err := storage.authorityID(ctx)
	if err != nil {
		return discountapp.EligibilitySnapshot{}, err
	}
	if err = storage.execute(ctx, lockRedemptionSQL, string(code)); err != nil {
		return discountapp.EligibilitySnapshot{}, fmt.Errorf(
			"lock billing discount redemption: %w", err,
		)
	}
	result := discountapp.EligibilitySnapshot{}
	var normalizedCode string
	var policySHA256 []byte
	err = storage.queryRow(ctx, lockEligibilitySQL, authorityID, storage.tenantID,
		string(code), subscriptionID, offeringVersionID, priceVersionID, now,
	).Scan(&result.DiscountVersionID, &normalizedCode, &result.OfferingVersionID,
		&result.PriceVersionID, &result.RequiredCurrency, &result.ValidFrom,
		&result.ValidUntil, &result.GlobalUseLimit, &result.GlobalUsed,
		&result.TenantUseLimit, &result.TenantUsed, &result.OfferingAllowed,
		&policySHA256)
	if err != nil {
		return discountapp.EligibilitySnapshot{}, fmt.Errorf(
			"lock billing discount eligibility: %w", err,
		)
	}
	normalized, err := discountdomain.NormalizeCode(normalizedCode)
	if err != nil || len(policySHA256) != len(result.PolicySHA256) {
		return discountapp.EligibilitySnapshot{}, discountapp.ErrSnapshot
	}
	result.Code = normalized
	copy(result.PolicySHA256[:], policySHA256)
	return result, nil
}

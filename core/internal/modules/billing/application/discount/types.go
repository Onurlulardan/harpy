package discount

import (
	"time"

	"github.com/google/uuid"

	discountdomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/discount"
)

type EligibilitySnapshot struct {
	DiscountVersionID uuid.UUID
	Code              discountdomain.Code
	OfferingVersionID uuid.UUID
	PriceVersionID    uuid.UUID
	RequiredCurrency  string
	ValidFrom         time.Time
	ValidUntil        time.Time
	GlobalUseLimit    *int64
	GlobalUsed        int64
	TenantUseLimit    *int64
	TenantUsed        int64
	OfferingAllowed   bool
	PolicySHA256      [32]byte
}

type Redemption struct {
	ID                uuid.UUID
	DiscountVersionID uuid.UUID
	SubscriptionID    uuid.UUID
	OfferingVersionID uuid.UUID
	PriceVersionID    uuid.UUID
	SourceOperationID uuid.UUID
	PolicySHA256      [32]byte
	RedeemedAt        time.Time
}

type RedeemRequest struct {
	Code              discountdomain.Code
	SubscriptionID    uuid.UUID
	OfferingVersionID uuid.UUID
	PriceVersionID    uuid.UUID
	Currency          string
	SourceOperationID uuid.UUID
}

type RedeemResult struct {
	Redemption Redemption
	Replay     bool
}

type RedeemExecution struct {
	Request RedeemRequest
	Now     time.Time
}

type RedemptionPlan struct {
	Result RedeemResult
	Create *Redemption
}

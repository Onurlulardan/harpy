package discount

import (
	"time"

	discountdomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/discount"
	"github.com/google/uuid"
)

type EligibilityInput struct {
	TenantID, DiscountID, DiscountVersionID uuid.UUID
	SubscriptionID, OfferingVersionID       uuid.UUID
	PriceVersionID                          uuid.UUID
	Revision, Version                       int64
	Currency                                string
	Policy                                  discountdomain.Policy
}

type Eligibility struct {
	tenantID, discountID, discountVersionID uuid.UUID
	subscriptionID, offeringVersionID       uuid.UUID
	priceVersionID                          uuid.UUID
	revision, version                       int64
	currency                                string
	policy                                  discountdomain.Policy
}

type EligibilityLookup struct {
	Code              discountdomain.Code
	OfferingVersionID uuid.UUID
	At                time.Time
}

func NewEligibility(input EligibilityInput) (Eligibility, error) {
	value := Eligibility{
		tenantID: input.TenantID, discountID: input.DiscountID,
		discountVersionID: input.DiscountVersionID,
		subscriptionID:    input.SubscriptionID, offeringVersionID: input.OfferingVersionID,
		priceVersionID: input.PriceVersionID, revision: input.Revision,
		version: input.Version, currency: input.Currency, policy: input.Policy,
	}
	if !value.valid() {
		return Eligibility{}, ErrQueryInvariant
	}
	return value, nil
}

func (value Eligibility) TenantID() uuid.UUID           { return value.tenantID }
func (value Eligibility) DiscountID() uuid.UUID         { return value.discountID }
func (value Eligibility) DiscountVersionID() uuid.UUID  { return value.discountVersionID }
func (value Eligibility) SubscriptionID() uuid.UUID     { return value.subscriptionID }
func (value Eligibility) OfferingVersionID() uuid.UUID  { return value.offeringVersionID }
func (value Eligibility) PriceVersionID() uuid.UUID     { return value.priceVersionID }
func (value Eligibility) Revision() int64               { return value.revision }
func (value Eligibility) Version() int64                { return value.version }
func (value Eligibility) Currency() string              { return value.currency }
func (value Eligibility) Policy() discountdomain.Policy { return value.policy }

func (value Eligibility) valid() bool {
	if value.tenantID == uuid.Nil || !validV7(value.discountID) ||
		!validV7(value.discountVersionID) || !validV7(value.subscriptionID) ||
		!validV7(value.offeringVersionID) || !validV7(value.priceVersionID) ||
		value.revision < 1 || value.version < 1 || !currencyPattern.MatchString(value.currency) ||
		!value.policy.Valid() {
		return false
	}
	if fixed, ok := value.policy.FixedAmount(); ok && fixed.Currency() != value.currency {
		return false
	}
	if value.policy.AllOfferings() {
		return len(value.policy.OfferingIDs()) == 0
	}
	for _, identifier := range value.policy.OfferingIDs() {
		if uuid.UUID(identifier) == value.offeringVersionID {
			return true
		}
	}
	return false
}

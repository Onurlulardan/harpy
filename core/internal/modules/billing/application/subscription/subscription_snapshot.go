package subscriptionapp

import (
	"time"

	"github.com/google/uuid"

	billingperiod "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/period"
	subscriptiondomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/subscription"
)

type SubscriptionSnapshot struct {
	ID                uuid.UUID
	TenantID          uuid.UUID
	ProviderAccountID uuid.UUID
	OfferingVersionID uuid.UUID
	PriceVersionID    uuid.UUID
	Currency          string
	State             subscriptiondomain.State
	Period            *billingperiod.Period
	TrialEndsAt       *time.Time
	GraceEndsAt       *time.Time
	CancelAtPeriodEnd bool
	CancelAt          *time.Time
	Revision          int64
	UpdatedAt         time.Time
}

func (snapshot SubscriptionSnapshot) valid() bool {
	if !validID(snapshot.ID) || !validID(snapshot.TenantID) ||
		!validID(snapshot.ProviderAccountID) || !validID(snapshot.OfferingVersionID) ||
		!validID(snapshot.PriceVersionID) || !currencyPattern.MatchString(snapshot.Currency) ||
		!snapshot.State.Valid() || snapshot.Revision < 1 || !validInstant(snapshot.UpdatedAt) ||
		(snapshot.CancelAtPeriodEnd != (snapshot.CancelAt != nil)) {
		return false
	}
	if snapshot.Period != nil && (!validInstant(snapshot.Period.Start()) ||
		!validInstant(snapshot.Period.End())) {
		return false
	}
	if snapshot.TrialEndsAt != nil && !validInstant(*snapshot.TrialEndsAt) ||
		snapshot.GraceEndsAt != nil && !validInstant(*snapshot.GraceEndsAt) ||
		snapshot.CancelAt != nil && !validInstant(*snapshot.CancelAt) {
		return false
	}
	if snapshot.CancelAt != nil && (snapshot.Period == nil ||
		!snapshot.CancelAt.Equal(snapshot.Period.End())) {
		return false
	}
	return snapshot.validStateWindow()
}

func (snapshot SubscriptionSnapshot) validStateWindow() bool {
	switch snapshot.State {
	case subscriptiondomain.Trialing:
		return snapshot.Period != nil && snapshot.TrialEndsAt != nil &&
			snapshot.TrialEndsAt.Equal(snapshot.Period.End())
	case subscriptiondomain.Active:
		return snapshot.Period != nil
	case subscriptiondomain.PastDue:
		return snapshot.GraceEndsAt != nil
	default:
		return true
	}
}

func (snapshot SubscriptionSnapshot) grantPeriod() (billingperiod.Period, bool) {
	if snapshot.Period == nil ||
		(snapshot.State != subscriptiondomain.Active && snapshot.State != subscriptiondomain.Trialing) {
		return billingperiod.Period{}, false
	}
	return *snapshot.Period, true
}

func (snapshot SubscriptionSnapshot) revised(now time.Time) SubscriptionSnapshot {
	copy := snapshot
	copy.Revision++
	copy.UpdatedAt = now
	copy.TrialEndsAt = copiedInstant(snapshot.TrialEndsAt)
	copy.GraceEndsAt = copiedInstant(snapshot.GraceEndsAt)
	copy.CancelAt = copiedInstant(snapshot.CancelAt)
	return copy
}

package subscriptionapp

import (
	"regexp"
	"time"

	billingperiod "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/period"
	subscriptiondomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/subscription"
	"github.com/google/uuid"
)

var providerKeyPattern = regexp.MustCompile(`^[a-z][a-z0-9._-]{0,62}$`)

type CurrentSubscriptionInput struct {
	ID, TenantID, ProviderAccountID   uuid.UUID
	LegacyID, LegacyPlanID            uuid.UUID
	OfferingVersionID, PriceVersionID uuid.UUID
	Provider                          string
	State                             subscriptiondomain.State
	Period                            *billingperiod.Period
	TrialEndsAt, GraceEndsAt          *time.Time
	TrialStartedAt, CancelledAt       *time.Time
	CancellationReason                string
	CancelAtPeriodEnd                 bool
	CancelAt                          *time.Time
	EffectiveMinor                    int64
	EffectiveCurrency                 string
	EffectiveRegistry                 int64
	Revision                          int64
	CreatedAt, UpdatedAt              time.Time
	Offering                          CurrentOffering
}

type CurrentSubscription struct {
	id, tenantID, providerAccountID   uuid.UUID
	legacyID, legacyPlanID            uuid.UUID
	offeringVersionID, priceVersionID uuid.UUID
	provider                          string
	state                             subscriptiondomain.State
	period                            *billingperiod.Period
	trialEndsAt, graceEndsAt          *time.Time
	trialStartedAt, cancelledAt       *time.Time
	cancellationReason                string
	cancelAtPeriodEnd                 bool
	cancelAt                          *time.Time
	effectiveMinor                    int64
	effectiveCurrency                 string
	effectiveRegistry                 int64
	revision                          int64
	createdAt, updatedAt              time.Time
	offering                          CurrentOffering
}

func NewCurrentSubscription(input CurrentSubscriptionInput) (CurrentSubscription, error) {
	value := CurrentSubscription{
		id: input.ID, tenantID: input.TenantID,
		legacyID: input.LegacyID, legacyPlanID: input.LegacyPlanID,
		providerAccountID: input.ProviderAccountID,
		offeringVersionID: input.OfferingVersionID,
		priceVersionID:    input.PriceVersionID, provider: input.Provider,
		state: input.State, period: copiedPeriod(input.Period),
		trialEndsAt:        copiedInstant(input.TrialEndsAt),
		graceEndsAt:        copiedInstant(input.GraceEndsAt),
		trialStartedAt:     copiedInstant(input.TrialStartedAt),
		cancelledAt:        copiedInstant(input.CancelledAt),
		cancellationReason: input.CancellationReason,
		cancelAtPeriodEnd:  input.CancelAtPeriodEnd,
		cancelAt:           copiedInstant(input.CancelAt), revision: input.Revision,
		effectiveMinor: input.EffectiveMinor, effectiveCurrency: input.EffectiveCurrency,
		effectiveRegistry: input.EffectiveRegistry,
		createdAt:         input.CreatedAt, updatedAt: input.UpdatedAt, offering: input.Offering,
	}
	if !value.valid() {
		return CurrentSubscription{}, ErrCurrentSubscriptionInvariant
	}
	return value, nil
}

func (value CurrentSubscription) ID() uuid.UUID                   { return value.id }
func (value CurrentSubscription) LegacyID() uuid.UUID             { return value.legacyID }
func (value CurrentSubscription) LegacyPlanID() uuid.UUID         { return value.legacyPlanID }
func (value CurrentSubscription) TenantID() uuid.UUID             { return value.tenantID }
func (value CurrentSubscription) ProviderAccountID() uuid.UUID    { return value.providerAccountID }
func (value CurrentSubscription) OfferingVersionID() uuid.UUID    { return value.offeringVersionID }
func (value CurrentSubscription) PriceVersionID() uuid.UUID       { return value.priceVersionID }
func (value CurrentSubscription) Provider() string                { return value.provider }
func (value CurrentSubscription) State() subscriptiondomain.State { return value.state }
func (value CurrentSubscription) EndOfPeriodCancellation() bool   { return value.cancelAtPeriodEnd }
func (value CurrentSubscription) Revision() int64                 { return value.revision }
func (value CurrentSubscription) EffectiveMinor() int64           { return value.effectiveMinor }
func (value CurrentSubscription) EffectiveCurrency() string       { return value.effectiveCurrency }
func (value CurrentSubscription) EffectiveRegistry() int64        { return value.effectiveRegistry }
func (value CurrentSubscription) CreatedAt() time.Time            { return value.createdAt }
func (value CurrentSubscription) UpdatedAt() time.Time            { return value.updatedAt }
func (value CurrentSubscription) Offering() CurrentOffering       { return value.offering }

func (value CurrentSubscription) Period() (billingperiod.Period, bool) {
	if value.period == nil {
		return billingperiod.Period{}, false
	}
	return *value.period, true
}

func (value CurrentSubscription) TrialEnd() (time.Time, bool) {
	return optionalInstant(value.trialEndsAt)
}
func (value CurrentSubscription) TrialStart() (time.Time, bool) {
	return optionalInstant(value.trialStartedAt)
}
func (value CurrentSubscription) Cancelled() (time.Time, bool) {
	return optionalInstant(value.cancelledAt)
}
func (value CurrentSubscription) CancellationReason() string {
	return value.cancellationReason
}
func (value CurrentSubscription) GraceEnd() (time.Time, bool) {
	return optionalInstant(value.graceEndsAt)
}
func (value CurrentSubscription) CancellationInstant() (time.Time, bool) {
	return optionalInstant(value.cancelAt)
}

func copiedPeriod(value *billingperiod.Period) *billingperiod.Period {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}

func optionalInstant(value *time.Time) (time.Time, bool) {
	if value == nil {
		return time.Time{}, false
	}
	return *value, true
}

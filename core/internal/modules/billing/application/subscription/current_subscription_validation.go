package subscriptionapp

import subscriptiondomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/subscription"

func (value CurrentSubscription) valid() bool {
	if !validID(value.id) || !validID(value.tenantID) ||
		!validID(value.legacyID) || !validID(value.legacyPlanID) ||
		!validID(value.providerAccountID) || !validID(value.offeringVersionID) ||
		!validID(value.priceVersionID) || !providerKeyPattern.MatchString(value.provider) ||
		value.effectiveMinor < 0 || value.effectiveRegistry < 1 ||
		!currencyPattern.MatchString(value.effectiveCurrency) || value.revision < 1 ||
		!validInstant(value.createdAt) || !validInstant(value.updatedAt) ||
		value.updatedAt.Before(value.createdAt) ||
		!value.offering.valid(len(value.offering.prices), len(value.offering.entitlements)) ||
		value.offering.VersionID() != value.offeringVersionID ||
		!value.offering.hasPrice(value.priceVersionID) ||
		value.cancelAtPeriodEnd != (value.cancelAt != nil) {
		return false
	}
	if value.state != subscriptiondomain.Trialing && value.state != subscriptiondomain.Active &&
		value.state != subscriptiondomain.PastDue {
		return false
	}
	if value.period != nil && (!validInstant(value.period.Start()) ||
		!validInstant(value.period.End())) {
		return false
	}
	if value.trialEndsAt != nil && !validInstant(*value.trialEndsAt) ||
		value.trialStartedAt != nil && !validInstant(*value.trialStartedAt) ||
		value.cancelledAt != nil && !validInstant(*value.cancelledAt) ||
		value.graceEndsAt != nil && !validInstant(*value.graceEndsAt) ||
		value.cancelAt != nil && !validInstant(*value.cancelAt) {
		return false
	}
	if value.cancelAt != nil && (value.period == nil ||
		!value.cancelAt.Equal(value.period.End())) {
		return false
	}
	switch value.state {
	case subscriptiondomain.Trialing:
		return value.period != nil && value.trialEndsAt != nil &&
			value.trialEndsAt.Equal(value.period.End())
	case subscriptiondomain.Active:
		return value.period != nil
	case subscriptiondomain.PastDue:
		return value.graceEndsAt != nil
	default:
		return false
	}
}

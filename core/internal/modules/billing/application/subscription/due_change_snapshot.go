package subscriptionapp

import (
	"time"

	subscriptiondomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/subscription"
)

func dueChangeSnapshot(
	before SubscriptionSnapshot,
	boundary BoundarySnapshot,
	now time.Time,
) (SubscriptionSnapshot, *ChangeResolution, error) {
	change, price, next := boundary.Change, boundary.TargetPrice, boundary.NextPeriod
	if change == nil || price == nil ||
		change.SubscriptionID != before.ID || !change.EffectiveAt.Equal(boundary.DueAt) ||
		change.BeforeOfferingVersionID != before.OfferingVersionID ||
		change.BeforePriceVersionID != before.PriceVersionID ||
		change.AfterOfferingVersionID != price.OfferingVersionID ||
		change.AfterPriceVersionID != price.PriceVersionID || price.Currency != before.Currency {
		return SubscriptionSnapshot{}, nil, ErrBoundaryMismatch
	}
	after := before.revised(now)
	after.OfferingVersionID, after.PriceVersionID = price.OfferingVersionID, price.PriceVersionID
	after.Currency = price.Currency
	after.CancelAtPeriodEnd, after.CancelAt = false, nil
	switch before.State {
	case subscriptiondomain.Active:
		if next == nil || boundary.PaymentReceiptID != nil || boundary.GraceEndsAt != nil ||
			!next.Start().Equal(boundary.DueAt) || !next.End().After(now) {
			return SubscriptionSnapshot{}, nil, ErrBoundaryMismatch
		}
		after.Period = next
	case subscriptiondomain.Trialing:
		if before.TrialEndsAt == nil || !before.TrialEndsAt.Equal(boundary.DueAt) {
			return SubscriptionSnapshot{}, nil, ErrBoundaryMismatch
		}
		after.TrialEndsAt = nil
		switch {
		case boundary.PaymentReceiptID != nil && next != nil &&
			next.Start().Equal(boundary.DueAt) && next.End().After(now):
			after.State, after.Period = subscriptiondomain.Active, next
			after.GraceEndsAt = nil
		case boundary.PaymentReceiptID == nil && next == nil &&
			boundary.GraceEndsAt != nil && boundary.GraceEndsAt.After(now):
			after.State = subscriptiondomain.PastDue
			after.GraceEndsAt = copiedInstant(boundary.GraceEndsAt)
		case boundary.PaymentReceiptID == nil && next == nil:
			after.State, after.GraceEndsAt = subscriptiondomain.Expired, nil
		default:
			return SubscriptionSnapshot{}, nil, ErrBoundaryMismatch
		}
		if subscriptiondomain.ValidateTransition(before.State, after.State) != nil {
			return SubscriptionSnapshot{}, nil, ErrBoundaryMismatch
		}
	default:
		return SubscriptionSnapshot{}, nil, ErrBoundaryMismatch
	}
	resolution := &ChangeResolution{
		ID: change.ID, ExpectedRevision: change.Revision,
		State: ChangeApplied, ResolvedAt: now,
	}
	return checkedDueSnapshot(after, resolution)
}

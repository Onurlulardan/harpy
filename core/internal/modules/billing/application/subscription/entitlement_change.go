package subscriptionapp

import "time"

func subscriptionEntitlementsChanged(
	before SubscriptionSnapshot,
	after SubscriptionSnapshot,
) bool {
	return before.OfferingVersionID != after.OfferingVersionID ||
		before.PriceVersionID != after.PriceVersionID ||
		before.State != after.State ||
		!sameGrantPeriod(before, after) ||
		!sameOptionalInstant(before.TrialEndsAt, after.TrialEndsAt) ||
		!sameOptionalInstant(before.GraceEndsAt, after.GraceEndsAt)
}

func sameGrantPeriod(left, right SubscriptionSnapshot) bool {
	if left.Period == nil || right.Period == nil {
		return left.Period == nil && right.Period == nil
	}
	return left.Period.Start().Equal(right.Period.Start()) &&
		left.Period.End().Equal(right.Period.End())
}

func sameOptionalInstant(left, right *time.Time) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return left.Equal(*right)
}

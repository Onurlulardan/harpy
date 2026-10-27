package subscriptionapp

import "github.com/google/uuid"

type TargetPrice struct {
	OfferingVersionID uuid.UUID
	PriceVersionID    uuid.UUID
	Currency          string
	TrialDays         int
}

func (target TargetPrice) valid() bool {
	return validID(target.OfferingVersionID) && validID(target.PriceVersionID) &&
		currencyPattern.MatchString(target.Currency) && target.TrialDays >= 0
}

package billingv1adoptionapp

import "github.com/google/uuid"

type OfferingIndex struct {
	SourceID          uuid.UUID
	OfferingID        string
	OfferingVersionID string
	MonthlyPriceID    string
	MonthlyMinorUnits int64
	Currency          string
}

type DiscountIndex struct {
	SourceID  uuid.UUID
	VersionID string
	Record    DiscountRecord
}

type GlobalIndex struct {
	Base                    OfferingIndex
	Offerings               map[uuid.UUID]OfferingIndex
	Discounts               map[uuid.UUID]DiscountIndex
	ManualProviderAccountID string
}

func newGlobalIndex() GlobalIndex {
	return GlobalIndex{
		Offerings:               make(map[uuid.UUID]OfferingIndex),
		Discounts:               make(map[uuid.UUID]DiscountIndex),
		ManualProviderAccountID: mustIdentity("provider-account", "manual", "default"),
	}
}

package httpapi

import "github.com/google/uuid"

type TenantSubscriptionSnapshot struct {
	ID, TenantID                      uuid.UUID
	OfferingVersionID, PriceVersionID uuid.UUID
	View                              SubscriptionView
}

type TenantUsageSnapshot struct {
	TenantID uuid.UUID
	View     UsageView
}

type TenantDiscountSnapshot struct {
	TenantID, DiscountID, DiscountVersionID uuid.UUID
	SubscriptionID, OfferingVersionID       uuid.UUID
	PriceVersionID                          uuid.UUID
	Currency                                string
	View                                    DiscountView
}

type TenantDiscountLookup struct {
	Code              string
	OfferingVersionID uuid.UUID
}

type TenantSubscriptionChange struct {
	ActorID, SubscriptionID, SourceOperationID    uuid.UUID
	TargetOfferingVersionID, TargetPriceVersionID uuid.UUID
	ExpectedRevision                              int64
}

type TenantSubscriptionCancellation struct {
	ActorID, SubscriptionID, SourceOperationID uuid.UUID
	ExpectedRevision                           int64
	Reason                                     string
}

type TenantSubscriptionResume struct {
	ActorID, SubscriptionID, SourceOperationID uuid.UUID
	ExpectedRevision                           int64
}

type TenantDiscountRedemption struct {
	Code                              string
	SubscriptionID, OfferingVersionID uuid.UUID
	PriceVersionID, SourceOperationID uuid.UUID
	Currency                          string
}

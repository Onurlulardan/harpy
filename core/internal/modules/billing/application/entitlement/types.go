package entitlement

import (
	"github.com/google/uuid"

	domain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/entitlement"
	billingperiod "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/period"
	"github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/subscription"
)

type PurchasedLayer struct {
	SubscriptionID    uuid.UUID
	OfferingVersionID uuid.UUID
	State             subscription.State
	GrantWindow       *billingperiod.Period
	Values            map[domain.Key]domain.Value
	Overrides         map[domain.Key]domain.Value
}

type Snapshot struct {
	Definitions           map[domain.Key]domain.Definition
	BaseAssignmentID      uuid.UUID
	BaseOfferingVersionID uuid.UUID
	Base                  map[domain.Key]domain.Value
	Purchased             *PurchasedLayer
}

type Resolution struct {
	values                map[domain.Key]domain.Value
	baseAssignmentID      uuid.UUID
	baseOfferingVersionID uuid.UUID
	subscriptionID        *uuid.UUID
	offeringVersionID     *uuid.UUID
}

func (resolution Resolution) Values() map[domain.Key]domain.Value {
	copy := make(map[domain.Key]domain.Value, len(resolution.values))
	for key, value := range resolution.values {
		copy[key] = value
	}
	return copy
}

func (resolution Resolution) BaseAssignmentID() uuid.UUID {
	return resolution.baseAssignmentID
}

func (resolution Resolution) BaseOfferingVersionID() uuid.UUID {
	return resolution.baseOfferingVersionID
}

func (resolution Resolution) SubscriptionID() *uuid.UUID {
	return copyIdentifier(resolution.subscriptionID)
}

func (resolution Resolution) OfferingVersionID() *uuid.UUID {
	return copyIdentifier(resolution.offeringVersionID)
}

func copyIdentifier(value *uuid.UUID) *uuid.UUID {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}

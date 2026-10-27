package discounteligibility

import (
	"context"
	"fmt"

	discountapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/discount"
	"github.com/google/uuid"
)

func (storage *store) FindEligible(
	ctx context.Context,
	lookup discountapp.EligibilityLookup,
) (discountapp.Eligibility, error) {
	if !storage.usable(ctx) || !lookup.Code.Valid() || !validInstant(lookup.At) ||
		lookup.OfferingVersionID != uuid.Nil && !validV7(lookup.OfferingVersionID) {
		return discountapp.Eligibility{}, discountapp.ErrQueryConfiguration
	}
	authority, err := storage.authorityIDs(ctx)
	if err != nil {
		return discountapp.Eligibility{}, err
	}
	record, err := scanEligibility(storage.queryRow(
		ctx, findEligibleSQL, authority.tenant, storage.tenantID,
		string(lookup.Code), nullableUUID(lookup.OfferingVersionID), lookup.At,
		authority.global,
	))
	if err != nil {
		if storage.notFound(err) {
			return discountapp.Eligibility{}, discountapp.ErrEligibilityMissing
		}
		return discountapp.Eligibility{}, fmt.Errorf(
			"%w: read discount eligibility", discountapp.ErrQueryStorage,
		)
	}
	if record.candidateCount != 1 {
		return discountapp.Eligibility{}, discountapp.ErrQueryInvariant
	}
	if !record.eligible() {
		return discountapp.Eligibility{}, discountapp.ErrEligibilityMissing
	}
	offerings, err := storage.listOfferings(
		ctx, authority.global, record.discountVersionID,
	)
	if err != nil {
		return discountapp.Eligibility{}, err
	}
	policy, err := record.policy(offerings)
	if err != nil {
		return discountapp.Eligibility{}, discountapp.ErrQueryInvariant
	}
	return discountapp.NewEligibility(discountapp.EligibilityInput{
		TenantID: storage.tenantID, DiscountID: record.discountID,
		DiscountVersionID: record.discountVersionID,
		SubscriptionID:    record.subscriptionID, OfferingVersionID: record.offeringID,
		PriceVersionID: record.priceID, Revision: record.revision,
		Version: record.version, Currency: record.subscriptionCurrency, Policy: policy,
	})
}

func nullableUUID(value uuid.UUID) any {
	if value == uuid.Nil {
		return nil
	}
	return value
}

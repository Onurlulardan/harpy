package billingv1adoptionapp

import (
	"math/big"
	"sort"

	"github.com/google/uuid"
)

func compileRedemptions(
	snapshot TenantSnapshot,
	global GlobalIndex,
	subscriptions map[uuid.UUID]SubscriptionRecord,
	issues []Issue,
) ([]RedemptionRecord, []Issue) {
	sourceSubscriptions := make(map[uuid.UUID]SourceSubscription, len(snapshot.Subscriptions))
	for _, source := range snapshot.Subscriptions {
		sourceSubscriptions[source.ID] = source
	}
	result := make([]RedemptionRecord, 0, len(snapshot.Redemptions))
	seenIDs, seenPolicy := map[uuid.UUID]struct{}{}, map[string]struct{}{}
	for _, source := range snapshot.Redemptions {
		key := source.ID.String()
		if source.ID == uuid.Nil || source.TenantID != snapshot.TenantID ||
			!canonicalInstant(source.AppliedAt) || !canonicalOptional(source.CreatedAt) ||
			source.DiscountAmountCents < 0 {
			issues = append(issues, issue("tenant", snapshot.TenantID.String(), "subscription_coupons", "snapshot_invalid", key))
			continue
		}
		if _, exists := seenIDs[source.ID]; exists {
			issues = append(issues, issue("tenant", snapshot.TenantID.String(), "subscription_coupons", "id_duplicate", key))
			continue
		}
		seenIDs[source.ID] = struct{}{}
		subscription, subscriptionExists := subscriptions[source.SubscriptionID]
		sourceSubscription, sourceExists := sourceSubscriptions[source.SubscriptionID]
		discount, discountExists := global.Discounts[source.CouponID]
		offering, offeringExists := global.Offerings[sourceSubscription.PlanID]
		if !subscriptionExists || !sourceExists || !discountExists || !offeringExists {
			issues = append(issues, issue("tenant", snapshot.TenantID.String(), "subscription_coupons", "relation_missing", key))
			continue
		}
		policyKey := subscription.ID + "\x00" + discount.VersionID
		if _, exists := seenPolicy[policyKey]; exists {
			issues = append(issues, issue("tenant", snapshot.TenantID.String(), "subscription_coupons", "subscription_policy_duplicate", key))
			continue
		}
		seenPolicy[policyKey] = struct{}{}
		record, err := convertRedemption(source, subscription, offering, discount.Record)
		if err != nil {
			issues = append(issues, issue("tenant", snapshot.TenantID.String(), "subscription_coupons", err.Error(), key))
			continue
		}
		result = append(result, record)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].SourceID < result[j].SourceID })
	return result, issues
}

func convertRedemption(
	source SourceRedemption,
	subscription SubscriptionRecord,
	offering OfferingIndex,
	discount DiscountRecord,
) (RedemptionRecord, error) {
	expected, err := legacyDiscountAmount(offering.MonthlyMinorUnits, discount)
	if err != nil || expected != source.DiscountAmountCents {
		return RedemptionRecord{}, adoptionCode("discount_amount_mismatch")
	}
	if discount.Duration == "repeating" && (source.RemainingMonths == nil ||
		discount.RepeatingCycles == nil || *source.RemainingMonths != *discount.RepeatingCycles) {
		return RedemptionRecord{}, adoptionCode("remaining_cycles_ambiguous")
	}
	if discount.Duration != "repeating" && source.RemainingMonths != nil {
		return RedemptionRecord{}, adoptionCode("remaining_cycles_invalid")
	}
	policySHA, err := recordDigest(discount)
	if err != nil {
		return RedemptionRecord{}, err
	}
	record := RedemptionRecord{SourceID: source.ID.String(),
		ID:                mustIdentity("redemption", source.ID.String()),
		SourceOperationID: mustIdentity("redemption-source-operation", source.ID.String()),
		TenantID:          source.TenantID.String(), SubscriptionID: subscription.ID,
		DiscountVersionID: discount.VersionID, OfferingVersionID: offering.OfferingVersionID,
		PriceVersionID: offering.MonthlyPriceID, AppliedKind: discount.DiscountKind,
		AppliedFixedMinorUnits: discount.FixedMinorUnits, AppliedCurrency: discount.Currency,
		AppliedBasisPoints: discount.BasisPoints, AppliedPriority: discount.Priority,
		AppliedDuration: discount.Duration, AppliedRepeatingCycles: discount.RepeatingCycles,
		SourceDiscountMinorUnits: source.DiscountAmountCents, PolicySHA256: policySHA,
		RedeemedAt: source.AppliedAt,
	}
	if subscription.State == "cancelled" || subscription.State == "expired" {
		if subscription.UpdatedAt.Before(source.AppliedAt) {
			return RedemptionRecord{}, adoptionCode("relationship_end_invalid")
		}
		value := subscription.UpdatedAt
		record.RelationshipEndedAt = &value
	}
	return record, nil
}

func legacyDiscountAmount(gross int64, discount DiscountRecord) (int64, error) {
	if discount.DiscountKind == "fixed" && discount.FixedMinorUnits != nil {
		if *discount.FixedMinorUnits > gross {
			return gross, nil
		}
		return *discount.FixedMinorUnits, nil
	}
	if discount.DiscountKind != "basis_point" || discount.BasisPoints == nil {
		return 0, adoptionCode("discount_policy_invalid")
	}
	product := new(big.Int).Mul(big.NewInt(gross), big.NewInt(int64(*discount.BasisPoints)))
	product.Quo(product, big.NewInt(10_000))
	if !product.IsInt64() {
		return 0, adoptionCode("discount_amount_overflow")
	}
	return product.Int64(), nil
}

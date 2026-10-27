package billingv1adoptionapp

import (
	"sort"

	"github.com/google/uuid"
)

func discountApplicability(
	source SourceCoupon,
	mapping GlobalMapping,
	index GlobalIndex,
	target *DiscountRecord,
) error {
	if len(source.ApplicablePlanIDs) == 0 {
		evidence := mapping.GlobalCouponEvidenceSHA256[source.ID]
		if !validDigest(evidence) {
			return adoptionCode("global_applicability_unproven")
		}
		target.AllOfferings = true
		target.OfferingVersionIDs = []string{}
		target.ApplicabilityEvidence = &evidence
		return nil
	}
	seen := make(map[uuid.UUID]struct{}, len(source.ApplicablePlanIDs))
	versions := make([]string, 0, len(source.ApplicablePlanIDs))
	for _, sourcePlanID := range source.ApplicablePlanIDs {
		if sourcePlanID == uuid.Nil {
			return adoptionCode("applicability_invalid")
		}
		if _, exists := seen[sourcePlanID]; exists {
			return adoptionCode("applicability_duplicate")
		}
		seen[sourcePlanID] = struct{}{}
		offering, exists := index.Offerings[sourcePlanID]
		if !exists {
			return adoptionCode("applicability_plan_missing")
		}
		versions = append(versions, offering.OfferingVersionID)
	}
	sort.Strings(versions)
	target.AllOfferings = false
	target.OfferingVersionIDs = versions
	return nil
}

func discountCurrencies(target *DiscountRecord, index GlobalIndex) []string {
	values := map[string]struct{}{}
	if target.AllOfferings {
		for _, offering := range index.Offerings {
			values[offering.Currency] = struct{}{}
		}
	} else {
		allowed := make(map[string]struct{}, len(target.OfferingVersionIDs))
		for _, id := range target.OfferingVersionIDs {
			allowed[id] = struct{}{}
		}
		for _, offering := range index.Offerings {
			if _, exists := allowed[offering.OfferingVersionID]; exists {
				values[offering.Currency] = struct{}{}
			}
		}
	}
	result := make([]string, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

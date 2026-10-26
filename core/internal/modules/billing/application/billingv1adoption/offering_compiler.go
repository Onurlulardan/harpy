package billingv1adoptionapp

import (
	"sort"

	"github.com/google/uuid"
)

func compileOfferings(
	sources []SourcePlan,
	mapping GlobalMapping,
	environmentID uuid.UUID,
	index GlobalIndex,
	issues []Issue,
) ([]OfferingRecord, GlobalIndex, []Issue) {
	result := make([]OfferingRecord, 0, len(sources))
	seenIDs, seenKeys := map[uuid.UUID]struct{}{}, map[string]struct{}{}
	for _, source := range sources {
		key := source.ID.String()
		if source.ID == uuid.Nil {
			issues = append(issues, issue("global", environmentID.String(), "plans", "id_missing", key))
			continue
		}
		if _, exists := seenIDs[source.ID]; exists {
			issues = append(issues, issue("global", environmentID.String(), "plans", "id_duplicate", key))
			continue
		}
		seenIDs[source.ID] = struct{}{}
		record, value, err := convertOffering(source, source.ID == mapping.BasePlanID)
		if err != nil {
			issues = append(issues, issue("global", environmentID.String(), "plans", err.Error(), key))
			continue
		}
		if _, exists := seenKeys[record.Key]; exists {
			issues = append(issues, issue("global", environmentID.String(), "plans", "normalized_key_duplicate", record.Key))
			continue
		}
		seenKeys[record.Key] = struct{}{}
		result = append(result, record)
		index.Offerings[source.ID] = value
	}
	sort.Slice(result, func(i, j int) bool { return result[i].SourceID < result[j].SourceID })
	return result, index, issues
}

func convertOffering(source SourcePlan, base bool) (OfferingRecord, OfferingIndex, error) {
	if source.CreatedAt == nil || !canonicalInstant(*source.CreatedAt) ||
		!canonicalOptional(source.UpdatedAt) || !canonicalOptional(source.DeletedAt) {
		return OfferingRecord{}, OfferingIndex{}, adoptionCode("timestamp_invalid")
	}
	updatedAt := *source.CreatedAt
	if source.UpdatedAt != nil {
		updatedAt = *source.UpdatedAt
	}
	if updatedAt.Before(*source.CreatedAt) {
		return OfferingRecord{}, OfferingIndex{}, adoptionCode("timestamp_invalid")
	}
	available := source.IsActive && source.DeletedAt == nil
	retiredAt, ok := offeringRetiredAt(source.DeletedAt, *source.CreatedAt)
	if !ok {
		return OfferingRecord{}, OfferingIndex{}, adoptionCode("lifecycle_ambiguous")
	}
	converted, err := ConvertPlan(ConversionRequest{Snapshot: PlanSnapshot{
		Name: source.Name, Slug: source.Slug, Description: source.Description,
		SortOrder: source.SortOrder, HighlightLabel: source.HighlightLabel,
		Color: source.Color, PriceCents: source.PriceCents,
		AnnualPriceCents: source.AnnualPriceCents, BillingInterval: source.BillingInterval,
		Currency: source.Currency, LimitsJSON: source.LimitsJSON, FeaturesJSON: source.FeaturesJSON,
		IsPublic: source.IsPublic, TrialDays: source.TrialDays,
	}, IsBase: base, Available: available, SnapshotAt: *source.CreatedAt})
	if err != nil || len(converted.MissingFields) != 0 {
		return OfferingRecord{}, OfferingIndex{}, adoptionCode("plan_conversion_blocked")
	}
	sourceID := source.ID.String()
	record := OfferingRecord{
		SourceID: sourceID, ID: mustIdentity("offering", sourceID),
		VersionID:              mustIdentity("offering-version", sourceID),
		PublicationOperationID: mustIdentity("offering-publication", sourceID),
		Key:                    string(converted.Key), Name: source.Name, Description: source.Description,
		SortOrder: source.SortOrder, HighlightLabel: source.HighlightLabel, Color: source.Color,
		BillingInterval: source.BillingInterval,
		IsBase:          base, IsPublic: source.IsPublic, Available: available,
		TrialDays: source.TrialDays, CreatedAt: *source.CreatedAt,
		UpdatedAt: updatedAt, RetiredAt: retiredAt,
	}
	record.Prices = offeringPrices(sourceID, converted)
	record.Grants = offeringGrants(converted)
	monthly := record.Prices[0]
	return record, OfferingIndex{SourceID: source.ID, OfferingID: record.ID,
		OfferingVersionID: record.VersionID, MonthlyPriceID: monthly.ID,
		MonthlyMinorUnits: monthly.MinorUnits, Currency: monthly.Currency}, nil
}

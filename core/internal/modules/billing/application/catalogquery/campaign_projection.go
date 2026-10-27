package catalogquery

import (
	"bytes"
	"time"

	campaigndomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/campaign"
	discountdomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/discount"
	"github.com/google/uuid"
)

func projectCampaign(record CampaignRecord, instantValidAtRecord bool) (CampaignSnapshot, bool) {
	key, err := campaigndomain.NewKey(record.Key)
	identifiers := make([]campaigndomain.Identifier, len(record.Discounts))
	for index, discount := range record.Discounts {
		identifiers[index] = campaigndomain.Identifier(discount.VersionID)
	}
	publication, publicationErr := campaigndomain.NewPublication(
		record.Name, record.Description, record.BannerText, record.AutoApply,
		record.Priority, record.StartsAt, record.EndsAt, identifiers,
	)
	if err != nil || record.ID == [16]byte{} || record.LegacyID == uuid.Nil ||
		record.VersionID == [16]byte{} ||
		record.Revision < 1 || record.Version < 1 || publicationErr != nil ||
		!validInstant(record.StartsAt) || !validInterval(record.StartsAt, record.EndsAt) ||
		!validInstant(record.CreatedAt) || !validInstant(record.UpdatedAt) ||
		record.UpdatedAt.Before(record.CreatedAt) ||
		!instantValidAtRecord {
		return CampaignSnapshot{}, false
	}
	discounts, valid := projectDiscounts(record.Discounts)
	if !valid {
		return CampaignSnapshot{}, false
	}
	return CampaignSnapshot{id: record.ID, legacyID: record.LegacyID,
		createdBy: record.CreatedBy, key: string(key), revision: record.Revision,
		versionID: record.VersionID, version: record.Version, state: "active",
		name: publication.Name(), description: publication.Description(),
		bannerText: publication.BannerText(), autoApply: publication.AutoApply(),
		available: record.Available, priority: publication.Priority(),
		startsAt: publication.StartsAt(), endsAt: publication.EndsAt(),
		createdAt: record.CreatedAt, updatedAt: record.UpdatedAt, discounts: discounts}, true
}

func projectDiscounts(records []DiscountRecord) ([]DiscountSnapshot, bool) {
	result := make([]DiscountSnapshot, len(records))
	for index, record := range records {
		if !validDiscount(record) || (index > 0 && compareDiscount(
			records[index-1], record,
		) >= 0) {
			return nil, false
		}
		state := record.State
		if state == "" {
			state = "active"
		}
		result[index] = DiscountSnapshot{id: record.ID, legacyID: record.LegacyID,
			createdBy: record.CreatedBy, code: record.Code,
			revision: record.Revision, versionID: record.VersionID, version: record.Version,
			state: state, name: record.Name, description: record.Description,
			available: record.Available,
			kind:      record.Kind, basisPoints: copyInt64(record.BasisPoints),
			fixedMinorUnits: copyInt64(record.FixedMinorUnits), currency: copyString(record.Currency),
			currencyRegistryVersion: copyInt64(record.CurrencyRegistryVersion),
			duration:                record.Duration, durationCycles: copyInt64(record.DurationCycles),
			priority: record.Priority, allOfferings: record.AllOfferings,
			offeringVersionIDs: append([]uuid.UUID(nil), record.OfferingVersionIDs...),
			stableOfferingIDs:  append([]uuid.UUID(nil), record.StableOfferingIDs...),
			legacyOfferingIDs:  append([]uuid.UUID(nil), record.LegacyOfferingIDs...),
			usedCount:          record.UsedCount,
			campaignIDs:        append([]uuid.UUID(nil), record.CampaignIDs...),
			legacyCampaignIDs:  append([]uuid.UUID(nil), record.LegacyCampaignIDs...),
			validStart:         record.ValidStart, validEnd: copyTime(record.ValidEnd),
			globalUseLimit:    copyInt64(record.GlobalUseLimit),
			perTenantUseLimit: copyInt64(record.PerTenantUseLimit),
			createdAt:         record.CreatedAt, updatedAt: record.UpdatedAt}
	}
	return result, true
}

func validDiscount(record DiscountRecord) bool {
	code, err := discountdomain.NormalizeCode(record.Code)
	presentation, presentationErr := discountdomain.NewPresentation(record.Name, record.Description)
	stateValid := record.State == "" || record.State == "active" || record.State == "revoked"
	return err == nil && string(code) == record.Code && record.ID != [16]byte{} &&
		record.LegacyID != uuid.Nil &&
		presentationErr == nil && presentation.Valid() &&
		record.VersionID != [16]byte{} && record.Revision > 0 && record.Version > 0 &&
		stateValid && record.Priority >= 0 && record.UsedCount >= 0 &&
		validRule(record) && validDuration(record) &&
		validLimit(record.GlobalUseLimit) && validLimit(record.PerTenantUseLimit) &&
		validInstant(record.ValidStart) && validInterval(record.ValidStart, record.ValidEnd) &&
		validInstant(record.CreatedAt) && validInstant(record.UpdatedAt) &&
		!record.UpdatedAt.Before(record.CreatedAt) &&
		validOfferingIDs(record.AllOfferings, record.OfferingVersionIDs) &&
		validOfferingIDs(record.AllOfferings, record.StableOfferingIDs) &&
		len(record.OfferingVersionIDs) == len(record.StableOfferingIDs) &&
		len(record.StableOfferingIDs) == len(record.LegacyOfferingIDs) &&
		validSortedIDs(record.CampaignIDs, 256) &&
		len(record.CampaignIDs) == len(record.LegacyCampaignIDs)
}

func validRule(record DiscountRecord) bool {
	if record.Kind == "basis_point" {
		return record.BasisPoints != nil && *record.BasisPoints >= 1 && *record.BasisPoints <= 10000 &&
			record.FixedMinorUnits == nil && record.Currency == nil &&
			record.CurrencyRegistryVersion == nil
	}
	return record.Kind == "fixed" && record.BasisPoints == nil &&
		record.FixedMinorUnits != nil && *record.FixedMinorUnits > 0 &&
		record.Currency != nil && validCurrency(*record.Currency) &&
		record.CurrencyRegistryVersion != nil && *record.CurrencyRegistryVersion > 0
}

func validDuration(record DiscountRecord) bool {
	return record.Duration == "repeating" && record.DurationCycles != nil &&
		*record.DurationCycles > 0 || (record.Duration == "once" || record.Duration == "forever") &&
		record.DurationCycles == nil
}

func validOfferingIDs(all bool, values []uuid.UUID) bool {
	return all == (len(values) == 0) && validSortedIDs(values, 256)
}

func validSortedIDs(values []uuid.UUID, maximum int) bool {
	if len(values) > maximum {
		return false
	}
	for index, value := range values {
		if value == [16]byte{} || index > 0 && bytes.Compare(values[index-1][:], value[:]) >= 0 {
			return false
		}
	}
	return true
}

func compareDiscount(left, right DiscountRecord) int {
	if left.Priority < right.Priority {
		return -1
	}
	if left.Priority > right.Priority {
		return 1
	}
	return bytes.Compare(left.VersionID[:], right.VersionID[:])
}

func validLimit(value *int64) bool { return value == nil || *value > 0 }

func validInterval(start time.Time, end *time.Time) bool {
	return end == nil || validInstant(*end) && start.Before(*end)
}

func activeAt(start time.Time, end *time.Time, instant time.Time) bool {
	return !instant.Before(start) && (end == nil || instant.Before(*end))
}

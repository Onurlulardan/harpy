package billingv1adoptionapp

import (
	"regexp"
	"sort"
	"strings"

	discountdomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/discount"
	"github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/money"
	"github.com/google/uuid"
)

var normalizedCouponPattern = regexp.MustCompile(`^[A-Z0-9][A-Z0-9_-]{0,63}$`)

func compileDiscounts(
	sources []SourceCoupon,
	mapping GlobalMapping,
	index GlobalIndex,
	issues []Issue,
	environmentID uuid.UUID,
) ([]DiscountRecord, GlobalIndex, []Issue) {
	result := make([]DiscountRecord, 0, len(sources))
	seenIDs, seenCodes := map[uuid.UUID]struct{}{}, map[string]struct{}{}
	for _, source := range sources {
		key := source.ID.String()
		if source.ID == uuid.Nil {
			issues = append(issues, issue("global", environmentID.String(), "coupons", "id_missing", key))
			continue
		}
		if _, exists := seenIDs[source.ID]; exists {
			issues = append(issues, issue("global", environmentID.String(), "coupons", "id_duplicate", key))
			continue
		}
		seenIDs[source.ID] = struct{}{}
		record, err := convertDiscount(source, mapping, index)
		if err != nil {
			issues = append(issues, issue("global", environmentID.String(), "coupons", err.Error(), key))
			continue
		}
		if _, exists := seenCodes[record.NormalizedCode]; exists {
			issues = append(issues, issue("global", environmentID.String(), "coupons", "normalized_code_duplicate", record.NormalizedCode))
			continue
		}
		seenCodes[record.NormalizedCode] = struct{}{}
		result = append(result, record)
		index.Discounts[source.ID] = DiscountIndex{SourceID: source.ID,
			VersionID: record.VersionID, Record: record}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].SourceID < result[j].SourceID })
	return result, index, issues
}

func convertDiscount(
	source SourceCoupon,
	mapping GlobalMapping,
	index GlobalIndex,
) (DiscountRecord, error) {
	if source.CreatedAt == nil || !canonicalInstant(*source.CreatedAt) ||
		!canonicalInstant(source.ValidFrom) || !canonicalOptional(source.ValidUntil) ||
		(source.ValidUntil != nil && !source.ValidFrom.Before(*source.ValidUntil)) ||
		!canonicalOptional(source.UpdatedAt) || !canonicalOptional(source.DeletedAt) {
		return DiscountRecord{}, adoptionCode("timestamp_invalid")
	}
	updatedAt, ok := sourceUpdatedAt(*source.CreatedAt, source.UpdatedAt)
	if !ok || source.DeletedAt != nil && source.DeletedAt.Before(*source.CreatedAt) {
		return DiscountRecord{}, adoptionCode("timestamp_invalid")
	}
	presentation, err := discountdomain.NewPresentation(source.Name, source.Description)
	if err != nil {
		return DiscountRecord{}, adoptionCode("presentation_invalid")
	}
	code := strings.ToUpper(source.Code)
	if !normalizedCouponPattern.MatchString(code) || source.UsedCount < 0 ||
		source.MaxUsesPerTenant < 1 || (source.MaxUses != nil && (*source.MaxUses < 1 || source.UsedCount > *source.MaxUses)) {
		return DiscountRecord{}, adoptionCode("policy_invalid")
	}
	record := DiscountRecord{SourceID: source.ID.String(), CreatedBy: source.CreatedBy.String(),
		ID:                     mustIdentity("discount", source.ID.String()),
		VersionID:              mustIdentity("discount-version", source.ID.String()),
		PublicationOperationID: mustIdentity("discount-publication", source.ID.String()),
		NormalizedCode:         code, Name: presentation.Name(),
		Description: presentation.Description(),
		Priority:    0, GlobalUseLimit: source.MaxUses,
		PerTenantUseLimit: &source.MaxUsesPerTenant, SourceUsedCount: source.UsedCount,
		ValidStart: source.ValidFrom, ValidEnd: source.ValidUntil,
		IsActive:  source.IsActive && source.DeletedAt == nil,
		CreatedAt: *source.CreatedAt, UpdatedAt: updatedAt, RevokedAt: source.DeletedAt,
	}
	if err := discountDuration(source, &record); err != nil {
		return DiscountRecord{}, err
	}
	if err := discountApplicability(source, mapping, index, &record); err != nil {
		return DiscountRecord{}, err
	}
	if err := discountAmount(source, index, &record); err != nil {
		return DiscountRecord{}, err
	}
	return record, nil
}

func discountDuration(source SourceCoupon, target *DiscountRecord) error {
	target.Duration = source.Duration
	switch source.Duration {
	case "once", "forever":
		if source.DurationMonths != nil {
			return adoptionCode("duration_invalid")
		}
	case "repeating":
		if source.DurationMonths == nil || *source.DurationMonths < 1 {
			return adoptionCode("duration_invalid")
		}
		value := *source.DurationMonths
		target.RepeatingCycles = &value
	default:
		return adoptionCode("duration_invalid")
	}
	return nil
}

func discountAmount(source SourceCoupon, index GlobalIndex, target *DiscountRecord) error {
	switch source.DiscountType {
	case "percentage":
		if source.DiscountValue < 1 || source.DiscountValue > 100 {
			return adoptionCode("amount_invalid")
		}
		value := int32(source.DiscountValue * 100)
		target.DiscountKind, target.BasisPoints = "basis_point", &value
	case "fixed_amount":
		if source.DiscountValue < 1 || source.DiscountValue > money.MaximumAbsoluteMinorUnits {
			return adoptionCode("amount_invalid")
		}
		currencies := discountCurrencies(target, index)
		if len(currencies) != 1 {
			return adoptionCode("fixed_currency_ambiguous")
		}
		value, currency := source.DiscountValue, currencies[0]
		target.DiscountKind, target.FixedMinorUnits, target.Currency = "fixed", &value, &currency
	default:
		return adoptionCode("amount_invalid")
	}
	return nil
}

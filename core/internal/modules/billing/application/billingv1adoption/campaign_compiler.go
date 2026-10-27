package billingv1adoptionapp

import (
	"sort"

	campaigndomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/campaign"
	offeringdomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/offering"
	"github.com/google/uuid"
)

func compileCampaigns(
	sources []SourceCampaign,
	coupons []SourceCoupon,
	index GlobalIndex,
	issues []Issue,
	environmentID uuid.UUID,
) ([]CampaignRecord, []Issue) {
	priorities, ambiguous := campaignPriorities(sources)
	discounts := campaignDiscounts(coupons, index)
	known := make(map[uuid.UUID]struct{}, len(sources))
	result := make([]CampaignRecord, 0, len(sources))
	for _, source := range sources {
		key := source.ID.String()
		if source.ID == uuid.Nil {
			issues = append(issues, issue("global", environmentID.String(), "campaigns", "id_missing", key))
			continue
		}
		if _, exists := known[source.ID]; exists {
			issues = append(issues, issue("global", environmentID.String(), "campaigns", "id_duplicate", key))
			continue
		}
		known[source.ID] = struct{}{}
		record, err := convertCampaign(source, priorities[source.ID], discounts[source.ID])
		if err != nil {
			issues = append(issues, issue("global", environmentID.String(), "campaigns", err.Error(), key))
			continue
		}
		if ambiguous[source.ID] {
			issues = append(issues, issue("global", environmentID.String(), "campaigns", "legacy_order_ambiguous", key))
		}
		result = append(result, record)
	}
	for _, coupon := range coupons {
		if coupon.CampaignID == nil {
			continue
		}
		if _, exists := known[*coupon.CampaignID]; !exists {
			issues = append(issues, issue("global", environmentID.String(), "campaigns",
				"coupon_campaign_missing", coupon.ID.String()))
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].SourceID < result[j].SourceID })
	return result, issues
}

func convertCampaign(
	source SourceCampaign,
	priority int32,
	discountVersionIDs []string,
) (CampaignRecord, error) {
	if source.CreatedAt == nil || !canonicalInstant(*source.CreatedAt) ||
		!canonicalInstant(source.StartsAt) || !canonicalOptional(source.EndsAt) ||
		(source.EndsAt != nil && !source.StartsAt.Before(*source.EndsAt)) ||
		!canonicalOptional(source.UpdatedAt) || !canonicalOptional(source.DeletedAt) {
		return CampaignRecord{}, adoptionCode("timestamp_invalid")
	}
	key, err := offeringdomain.NewKey(source.Slug)
	if err != nil {
		return CampaignRecord{}, adoptionCode("key_invalid")
	}
	updatedAt, ok := sourceUpdatedAt(*source.CreatedAt, source.UpdatedAt)
	if !ok || source.DeletedAt != nil && source.DeletedAt.Before(*source.CreatedAt) {
		return CampaignRecord{}, adoptionCode("timestamp_invalid")
	}
	identifiers, ok := campaignIdentifiers(discountVersionIDs)
	if !ok {
		return CampaignRecord{}, adoptionCode("discount_identity_invalid")
	}
	publication, err := campaigndomain.NewPublication(
		source.Name, source.Description, source.BannerText, source.AutoApply,
		priority, source.StartsAt, source.EndsAt, identifiers,
	)
	if err != nil {
		return CampaignRecord{}, adoptionCode("publication_invalid")
	}
	return CampaignRecord{SourceID: source.ID.String(), CreatedBy: source.CreatedBy.String(),
		ID:                     mustIdentity("campaign", source.ID.String()),
		VersionID:              mustIdentity("campaign-version", source.ID.String()),
		PublicationOperationID: mustIdentity("campaign-publication", source.ID.String()),
		Key:                    string(key), Name: publication.Name(), Description: publication.Description(),
		BannerText: publication.BannerText(), AutoApply: publication.AutoApply(),
		Priority: publication.Priority(), IsActive: source.IsActive && source.DeletedAt == nil,
		StartsAt: publication.StartsAt(), EndsAt: publication.EndsAt(),
		CreatedAt: *source.CreatedAt, UpdatedAt: updatedAt, RetiredAt: source.DeletedAt,
		DiscountVersionIDs: discountVersionIDs,
	}, nil
}

func campaignIdentifiers(values []string) ([]campaigndomain.Identifier, bool) {
	result := make([]campaigndomain.Identifier, len(values))
	for index, value := range values {
		identifier, err := uuid.Parse(value)
		if err != nil || identifier == uuid.Nil || identifier.Version() != uuid.Version(7) {
			return nil, false
		}
		result[index] = campaigndomain.Identifier(identifier)
	}
	return result, true
}

func campaignDiscounts(coupons []SourceCoupon, index GlobalIndex) map[uuid.UUID][]string {
	result := map[uuid.UUID][]string{}
	for _, coupon := range coupons {
		if coupon.CampaignID == nil {
			continue
		}
		discount, exists := index.Discounts[coupon.ID]
		if !exists {
			continue
		}
		result[*coupon.CampaignID] = append(result[*coupon.CampaignID], discount.VersionID)
	}
	for campaignID := range result {
		sort.Strings(result[campaignID])
	}
	return result
}

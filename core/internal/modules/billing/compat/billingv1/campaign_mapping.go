package billingv1

import (
	"time"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func mapCampaignCreate(
	access billingapi.Access,
	key string,
	input campaignRequest,
) (billingapi.CreateCampaignCommand, error) {
	if input.Name == nil || input.Slug == nil || input.StartDate == nil ||
		!required(*input.Name) || !required(*input.Slug) {
		return billingapi.CreateCampaignCommand{}, ErrRequestInvalid
	}
	return billingapi.CreateCampaignCommand{
		Access: access, IdempotencyKey: key, Key: *input.Slug,
		InitialVersion: billingapi.CampaignVersionInput{
			Name: *input.Name, Description: stringValue(input.Description, ""),
			BannerText: stringValue(input.BannerText, ""),
			AutoApply:  boolValue(input.AutoApply, false),
			Available:  input.IsActive,
			StartsAt:   utcMicrosecond(*input.StartDate),
			EndsAt:     normalizeOptionalTime(input.EndDate.Value),
		},
	}, nil
}

func mapCampaignUpdate(
	access billingapi.Access,
	key string,
	current billingapi.CampaignView,
	input campaignRequest,
	now time.Time,
) (billingapi.CreateCampaignVersionCommand, error) {
	if input.Slug != nil && *input.Slug != current.Key {
		return billingapi.CreateCampaignVersionCommand{}, ErrUnsupported
	}
	name := current.Name
	if input.Name != nil {
		name = *input.Name
	}
	banner := current.BannerText
	if input.BannerText != nil {
		banner = *input.BannerText
	}
	description := current.Description
	if input.Description != nil {
		description = *input.Description
	}
	autoApply := current.AutoApply
	if input.AutoApply != nil {
		autoApply = *input.AutoApply
	}
	available := current.Available
	if input.IsActive != nil {
		available = *input.IsActive
	}
	startsAt := current.StartsAt
	if input.StartDate != nil {
		startsAt = utcMicrosecond(*input.StartDate)
	}
	endsAt := current.EndsAt
	if input.EndDate.Present {
		endsAt = normalizeOptionalTime(input.EndDate.Value)
	}
	discounts := make([]string, len(current.Discounts))
	for index, discount := range current.Discounts {
		discounts[index] = discount.VersionID
	}
	_ = now
	return billingapi.CreateCampaignVersionCommand{
		Access: access, IdempotencyKey: key, CampaignID: current.ID,
		ExpectedRevision: current.Revision,
		Version: billingapi.CampaignVersionInput{
			Name: name, Description: description, BannerText: banner,
			AutoApply: autoApply, Available: &available, Priority: current.Priority,
			StartsAt: startsAt, EndsAt: endsAt, DiscountVersionIDs: discounts,
		},
	}, nil
}

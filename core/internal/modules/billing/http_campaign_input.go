package billing

import billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"

type campaignInput struct {
	Key            string               `json:"key"`
	InitialVersion campaignVersionInput `json:"initial_version"`
}

type campaignVersionInput struct {
	ExpectedRevision   int64    `json:"expected_revision,omitempty"`
	Name               string   `json:"name"`
	BannerText         string   `json:"banner_text"`
	Priority           int      `json:"priority"`
	StartsAt           string   `json:"starts_at"`
	EndsAt             *string  `json:"ends_at,omitempty"`
	DiscountVersionIDs []string `json:"discount_version_ids"`
}

func (input campaignVersionInput) application() billingapi.CampaignVersionInput {
	return billingapi.CampaignVersionInput{
		Name: input.Name, BannerText: input.BannerText, Priority: input.Priority,
		StartsAt: input.StartsAt, EndsAt: input.EndsAt,
		DiscountVersionIDs: append([]string(nil), input.DiscountVersionIDs...),
	}
}

func (input campaignVersionInput) validInitial() bool {
	return input.ExpectedRevision == 0 && input.validPolicy()
}

func (input campaignVersionInput) validMutation() bool {
	return input.ExpectedRevision > 0 && input.validPolicy()
}

func (input campaignVersionInput) validPolicy() bool {
	return requiredText(input.Name, 128) && len(input.BannerText) <= 512 &&
		validUTCInterval(input.StartsAt, input.EndsAt) &&
		validUUIDv7Set(input.DiscountVersionIDs, 0, 256)
}

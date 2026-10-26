package httpapi

type CampaignView struct {
	ID          string         `json:"id"`
	LegacyID    string         `json:"legacy_id"`
	CreatedBy   string         `json:"created_by"`
	Key         string         `json:"key"`
	Revision    int64          `json:"revision"`
	VersionID   string         `json:"version_id"`
	Version     int64          `json:"version"`
	State       string         `json:"state"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	BannerText  string         `json:"banner_text"`
	AutoApply   bool           `json:"auto_apply"`
	Available   bool           `json:"available"`
	Priority    int            `json:"priority"`
	StartsAt    string         `json:"starts_at"`
	EndsAt      *string        `json:"ends_at,omitempty"`
	Discounts   []DiscountView `json:"discounts"`
	CreatedAt   string         `json:"created_at"`
	UpdatedAt   string         `json:"updated_at"`
}

type CampaignVersionInput struct {
	Name               string
	Description        string
	BannerText         string
	AutoApply          bool
	Available          *bool
	Priority           int
	StartsAt           string
	EndsAt             *string
	DiscountVersionIDs []string
}

type CreateCampaignCommand struct {
	Access         Access
	IdempotencyKey string
	Key            string
	InitialVersion CampaignVersionInput
}

type CreateCampaignVersionCommand struct {
	Access           Access
	IdempotencyKey   string
	CampaignID       string
	ExpectedRevision int64
	Version          CampaignVersionInput
}

type RetireCampaignCommand struct {
	Access           Access
	IdempotencyKey   string
	CampaignID       string
	ExpectedRevision int64
	Reason           string
}

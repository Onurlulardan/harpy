package billingv1

import billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"

type campaignWire struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Slug        string       `json:"slug"`
	Description string       `json:"description,omitempty"`
	StartDate   string       `json:"start_date"`
	EndDate     *string      `json:"end_date,omitempty"`
	AutoApply   bool         `json:"auto_apply"`
	BannerText  string       `json:"banner_text,omitempty"`
	IsActive    bool         `json:"is_active"`
	CreatedBy   string       `json:"created_by"`
	Coupons     []couponWire `json:"coupons,omitempty"`
	CreatedAt   string       `json:"created_at"`
	UpdatedAt   string       `json:"updated_at"`
}

func legacyCampaign(view billingapi.CampaignView) (campaignWire, error) {
	start, err := v1WireInstant(view.StartsAt)
	if err != nil {
		return campaignWire{}, ErrRequestInvalid
	}
	end, err := optionalV1WireInstant(view.EndsAt)
	if err != nil {
		return campaignWire{}, ErrRequestInvalid
	}
	created, err := v1WireInstant(view.CreatedAt)
	if err != nil {
		return campaignWire{}, ErrRequestInvalid
	}
	updated, err := v1WireInstant(view.UpdatedAt)
	if err != nil {
		return campaignWire{}, ErrRequestInvalid
	}
	identifier := view.LegacyID
	if identifier == "" {
		identifier = view.ID
	}
	return campaignWire{
		ID: identifier, Name: view.Name, Slug: view.Key,
		Description: view.Description, StartDate: start, EndDate: end,
		AutoApply: view.AutoApply, BannerText: view.BannerText,
		IsActive:  view.State == "active" && view.Available,
		CreatedBy: view.CreatedBy, CreatedAt: created, UpdatedAt: updated,
	}, nil
}

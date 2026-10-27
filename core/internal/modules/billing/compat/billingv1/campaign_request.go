package billingv1

import "time"

type campaignRequest struct {
	Name        *string      `json:"name"`
	Slug        *string      `json:"slug"`
	Description *string      `json:"description"`
	StartDate   *time.Time   `json:"start_date"`
	EndDate     optionalTime `json:"end_date"`
	AutoApply   *bool        `json:"auto_apply"`
	BannerText  *string      `json:"banner_text"`
	IsActive    *bool        `json:"is_active"`
}

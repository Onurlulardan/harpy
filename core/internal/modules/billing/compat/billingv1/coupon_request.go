package billingv1

import "time"

type couponRequest struct {
	Code              *string        `json:"code"`
	Name              *string        `json:"name"`
	Description       *string        `json:"description"`
	DiscountType      *string        `json:"discount_type"`
	DiscountValue     *int64         `json:"discount_value"`
	MaxUses           optionalInt64  `json:"max_uses"`
	MaxUsesPerTenant  *int64         `json:"max_uses_per_tenant"`
	ValidFrom         *time.Time     `json:"valid_from"`
	ValidUntil        optionalTime   `json:"valid_until"`
	ApplicablePlanIDs *[]string      `json:"applicable_plan_ids"`
	Duration          *string        `json:"duration"`
	DurationMonths    optionalInt    `json:"duration_months"`
	IsActive          *bool          `json:"is_active"`
	CampaignID        optionalString `json:"campaign_id"`
}

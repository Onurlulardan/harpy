package billingv1

import billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"

type couponWire struct {
	ID                string        `json:"id"`
	Code              string        `json:"code"`
	Name              string        `json:"name"`
	Description       string        `json:"description,omitempty"`
	DiscountType      string        `json:"discount_type"`
	DiscountValue     int64         `json:"discount_value"`
	MaxUses           *int64        `json:"max_uses"`
	UsedCount         int64         `json:"used_count"`
	MaxUsesPerTenant  int64         `json:"max_uses_per_tenant"`
	ValidFrom         string        `json:"valid_from"`
	ValidUntil        *string       `json:"valid_until,omitempty"`
	ApplicablePlanIDs []string      `json:"applicable_plan_ids"`
	Duration          string        `json:"duration"`
	DurationMonths    *int          `json:"duration_months,omitempty"`
	IsActive          bool          `json:"is_active"`
	CampaignID        *string       `json:"campaign_id,omitempty"`
	Campaign          *campaignWire `json:"campaign,omitempty"`
	CreatedBy         string        `json:"created_by"`
	CreatedAt         string        `json:"created_at"`
	UpdatedAt         string        `json:"updated_at"`
}

func legacyCoupon(view billingapi.DiscountView) (couponWire, error) {
	value, kind := int64(0), "percentage"
	if view.BasisPoints != nil {
		value = *view.BasisPoints / 100
	} else if view.FixedAmount != nil {
		kind, value = "fixed_amount", view.FixedAmount.MinorUnits
	}
	perTenant := int64(1)
	if view.PerTenantUseLimit != nil {
		perTenant = *view.PerTenantUseLimit
	}
	validFrom, err := v1WireInstant(view.ValidStart)
	if err != nil {
		return couponWire{}, ErrRequestInvalid
	}
	validUntil, err := optionalV1WireInstant(view.ValidEnd)
	if err != nil {
		return couponWire{}, ErrRequestInvalid
	}
	createdAt, err := v1WireInstant(view.CreatedAt)
	if err != nil {
		return couponWire{}, ErrRequestInvalid
	}
	updatedAt, err := v1WireInstant(view.UpdatedAt)
	if err != nil || view.AssociationAmbiguous {
		return couponWire{}, ErrRequestInvalid
	}
	identifier := view.LegacyID
	if identifier == "" {
		identifier = view.ID
	}
	return couponWire{
		ID: identifier, Code: view.Code, Name: view.Name, Description: view.Description,
		DiscountType: kind, DiscountValue: value,
		MaxUses: view.GlobalUseLimit, UsedCount: view.UsedCount,
		MaxUsesPerTenant: perTenant, ValidFrom: validFrom, ValidUntil: validUntil,
		ApplicablePlanIDs: append([]string{}, view.LegacyOfferingIDs...),
		Duration:          view.Duration, DurationMonths: view.DurationCycles,
		IsActive:   view.State == "active" && view.Available,
		CampaignID: view.LegacyCampaignID, CreatedBy: view.CreatedBy,
		CreatedAt: createdAt, UpdatedAt: updatedAt,
	}, nil
}

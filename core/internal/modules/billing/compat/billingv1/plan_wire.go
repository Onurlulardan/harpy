package billingv1

import billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"

type planWire struct {
	ID               string       `json:"id"`
	Name             string       `json:"name"`
	Slug             string       `json:"slug"`
	Description      string       `json:"description"`
	PriceCents       int64        `json:"price_cents"`
	AnnualPriceCents int64        `json:"annual_price_cents"`
	BillingInterval  string       `json:"billing_interval"`
	Currency         string       `json:"currency"`
	Limits           planLimits   `json:"limits"`
	Features         planFeatures `json:"features"`
	SortOrder        int          `json:"sort_order"`
	IsActive         bool         `json:"is_active"`
	IsPublic         bool         `json:"is_public"`
	HighlightLabel   string       `json:"highlight_label,omitempty"`
	Color            string       `json:"color,omitempty"`
	TrialDays        int          `json:"trial_days"`
	CreatedAt        string       `json:"created_at"`
	UpdatedAt        string       `json:"updated_at"`
}

func legacyPlan(offering billingapi.OfferingView) planWire {
	limits, features := legacyPlanValues(offering)
	plan := planWire{
		ID: offering.ID, Name: offering.Name, Slug: offering.Key,
		Description: offering.Description, BillingInterval: offering.BillingInterval,
		Limits: limits, Features: features, SortOrder: int(offering.SortOrder),
		HighlightLabel: offering.HighlightLabel, Color: offering.Color,
		IsActive: offering.State == "active" && offering.Available,
		IsPublic: offering.Public, TrialDays: offering.TrialDays,
		CreatedAt: v1JSONInstant(offering.CreatedAt), UpdatedAt: v1JSONInstant(offering.UpdatedAt),
	}
	for _, price := range offering.Prices {
		if plan.Currency == "" {
			plan.Currency = price.Amount.Currency
		}
		switch price.Cadence {
		case "monthly":
			plan.PriceCents = price.Amount.MinorUnits
		case "yearly":
			plan.AnnualPriceCents = price.Amount.MinorUnits
		}
	}
	return plan
}

func legacyPlanValues(
	offering billingapi.OfferingView,
) (planLimits, planFeatures) {
	var limits planLimits
	var features planFeatures
	for _, entitlement := range offering.Entitlements {
		integer := int64(0)
		if entitlement.Unlimited {
			integer = -1
		} else if entitlement.Integer != nil {
			integer = *entitlement.Integer
		}
		boolean := entitlement.Boolean != nil && *entitlement.Boolean
		switch entitlement.Key {
		case "forms.submissions":
			limits.SubmissionsPerMonth = integer
		case "iam.active_users":
			limits.MaxUsers = integer
		case "forms.active_forms":
			limits.MaxForms = integer
		case "documents.storage_bytes":
			if integer == -1 {
				limits.MaxStorageMB = -1
			} else {
				limits.MaxStorageMB = integer / (1024 * 1024)
			}
		case "ai.chat":
			features.AIChat = boolean
		case "workflow.enabled":
			features.Workflows = boolean
		case "api.access":
			features.APIAccess = boolean
		}
	}
	return limits, features
}

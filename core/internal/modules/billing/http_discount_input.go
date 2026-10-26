package billing

import billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"

type discountInput struct {
	InitialVersion discountVersionInput `json:"initial_version"`
}

type discountVersionInput struct {
	ExpectedRevision   int64                 `json:"expected_revision,omitempty"`
	Code               string                `json:"code"`
	Kind               string                `json:"kind"`
	BasisPoints        *int64                `json:"basis_points,omitempty"`
	FixedAmount        *billingapi.MoneyView `json:"fixed_amount,omitempty"`
	Duration           string                `json:"duration"`
	DurationCycles     *int                  `json:"duration_cycles,omitempty"`
	Priority           int                   `json:"priority"`
	AllOfferings       bool                  `json:"all_offerings"`
	OfferingVersionIDs []string              `json:"offering_version_ids"`
	ValidStart         string                `json:"valid_start"`
	ValidEnd           *string               `json:"valid_end,omitempty"`
	GlobalUseLimit     *int64                `json:"global_use_limit,omitempty"`
	PerTenantUseLimit  *int64                `json:"per_tenant_use_limit,omitempty"`
}

func (input discountVersionInput) application() billingapi.DiscountVersionInput {
	return billingapi.DiscountVersionInput{
		Code: input.Code, Kind: input.Kind, BasisPoints: input.BasisPoints,
		FixedAmount: input.FixedAmount, Duration: input.Duration,
		DurationCycles: input.DurationCycles, Priority: input.Priority,
		AllOfferings:       input.AllOfferings,
		OfferingVersionIDs: append([]string(nil), input.OfferingVersionIDs...),
		ValidStart:         input.ValidStart, ValidEnd: input.ValidEnd,
		GlobalUseLimit: input.GlobalUseLimit, PerTenantUseLimit: input.PerTenantUseLimit,
	}
}

func (input discountVersionInput) validInitial() bool {
	return input.ExpectedRevision == 0 && input.validPolicy()
}

func (input discountVersionInput) validMutation() bool {
	return input.ExpectedRevision > 0 && input.validPolicy()
}

func (input discountVersionInput) validPolicy() bool {
	return requiredText(input.Code, 64) &&
		input.validValue() && input.validDuration() &&
		validUTCInterval(input.ValidStart, input.ValidEnd) &&
		positiveOptional(input.GlobalUseLimit) && positiveOptional(input.PerTenantUseLimit) &&
		validOfferingEligibility(input.AllOfferings, input.OfferingVersionIDs)
}

func (input discountVersionInput) validValue() bool {
	if input.Kind == "basis_points" {
		return input.BasisPoints != nil && *input.BasisPoints > 0 &&
			*input.BasisPoints <= 10000 && input.FixedAmount == nil
	}
	return input.Kind == "fixed" && input.FixedAmount != nil && input.BasisPoints == nil &&
		input.FixedAmount.MinorUnits > 0 && input.FixedAmount.RegistryVersion > 0 &&
		validCurrency(input.FixedAmount.Currency)
}

func (input discountVersionInput) validDuration() bool {
	if input.Duration == "repeating" {
		return input.DurationCycles != nil && *input.DurationCycles > 0
	}
	return (input.Duration == "once" || input.Duration == "forever") &&
		input.DurationCycles == nil
}

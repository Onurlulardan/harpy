package billing

import (
	discountapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/discountcatalog"
	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	"github.com/google/uuid"
)

func globalHTTPDiscountInput(
	input billingapi.DiscountVersionInput,
) (discountapp.PolicyInput, error) {
	priority, priorityErr := int32Value(input.Priority)
	start, startErr := parseTimestamp(input.ValidStart)
	end, endErr := parseOptionalTimestamp(input.ValidEnd)
	if priorityErr != nil || startErr != nil || endErr != nil {
		return discountapp.PolicyInput{}, errInvalidInput
	}
	var fixed *discountapp.MoneyInput
	if input.FixedAmount != nil {
		fixed = &discountapp.MoneyInput{
			MinorUnits: input.FixedAmount.MinorUnits, Currency: input.FixedAmount.Currency,
			RegistryVersion: input.FixedAmount.RegistryVersion,
		}
	}
	return discountapp.PolicyInput{
		Name: input.Name, Description: input.Description,
		Code: input.Code, Kind: input.Kind, BasisPoints: input.BasisPoints,
		FixedAmount: fixed, Duration: input.Duration, DurationCycles: input.DurationCycles,
		Priority: priority, AllOfferings: input.AllOfferings,
		OfferingVersionIDs: append([]string(nil), input.OfferingVersionIDs...),
		ValidStart:         start, ValidEnd: end, GlobalUseLimit: input.GlobalUseLimit,
		PerTenantUseLimit: input.PerTenantUseLimit,
	}, nil
}

func globalHTTPCampaignUpdate(
	input billingapi.CampaignBindingInput,
) (discountapp.CampaignUpdate, error) {
	result := discountapp.CampaignUpdate{Replace: input.Replace}
	if input.CampaignID == nil {
		return result, nil
	}
	identifier, err := uuid.Parse(*input.CampaignID)
	if err != nil || identifier == uuid.Nil || identifier.Version() != uuid.Version(7) {
		return discountapp.CampaignUpdate{}, errInvalidInput
	}
	result.CampaignID = &identifier
	return result, nil
}

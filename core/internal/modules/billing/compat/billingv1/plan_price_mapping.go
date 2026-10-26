package billingv1

import (
	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	"github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/money"
)

func mergePlanPrices(
	current billingapi.OfferingView,
	input planRequest,
) ([]billingapi.PriceInput, error) {
	interval := current.BillingInterval
	if input.BillingInterval != nil {
		interval = *input.BillingInterval
	}
	if interval != "monthly" && interval != "annual" {
		return nil, ErrRequestInvalid
	}
	monthly, err := monthlyPrice(current)
	if err != nil {
		return nil, err
	}
	monthlyAmount := monthly.Amount.MinorUnits
	if input.PriceCents != nil {
		monthlyAmount = *input.PriceCents
	}
	currency := monthly.Amount.Currency
	registry := monthly.Amount.RegistryVersion
	if input.Currency != nil {
		currency, registry = *input.Currency, money.RegistryVersion
	}
	if monthlyAmount < 0 || currency == "" {
		return nil, ErrRequestInvalid
	}
	prices := []billingapi.PriceInput{{
		Cadence: "monthly", MinorUnits: monthlyAmount,
		Currency: currency, RegistryVersion: registry,
	}}
	annualAmount, annualRegistry := int64(0), registry
	for _, price := range current.Prices {
		if price.Cadence == "yearly" {
			annualAmount = price.Amount.MinorUnits
			annualRegistry = price.Amount.RegistryVersion
		}
	}
	if input.AnnualPriceCents != nil {
		annualAmount = *input.AnnualPriceCents
	}
	if input.Currency != nil {
		annualRegistry = money.RegistryVersion
	}
	if annualAmount < 0 {
		return nil, ErrRequestInvalid
	}
	if annualAmount > 0 {
		prices = append(prices, billingapi.PriceInput{
			Cadence: "yearly", MinorUnits: annualAmount,
			Currency: currency, RegistryVersion: annualRegistry,
		})
	}
	if interval == "annual" && annualAmount <= 0 {
		return nil, ErrRequestInvalid
	}
	return prices, nil
}

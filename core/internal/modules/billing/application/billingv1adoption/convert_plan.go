package billingv1adoptionapp

import (
	"strings"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/money"
	offeringdomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/offering"
)

func ConvertPlan(request ConversionRequest) (PlanResult, error) {
	plan := request.Snapshot
	if (plan.BillingInterval != "monthly" && plan.BillingInterval != "annual") ||
		(plan.BillingInterval == "annual" && plan.AnnualPriceCents <= 0) ||
		plan.PriceCents < 0 ||
		plan.AnnualPriceCents < 0 || plan.TrialDays < 0 || plan.TrialDays > 3650 ||
		request.SnapshotAt.IsZero() || request.SnapshotAt.Location() != time.UTC ||
		request.SnapshotAt.Nanosecond()%1_000 != 0 || strings.TrimSpace(plan.Name) != plan.Name {
		return PlanResult{}, ErrPlan
	}
	key, err := offeringdomain.NewKey(plan.Slug)
	if err != nil {
		return PlanResult{}, ErrPlan
	}
	limits, missingLimits, err := decodeLimits(plan.LimitsJSON)
	if err != nil {
		return PlanResult{}, err
	}
	features, missingFeatures, err := decodeFeatures(plan.FeaturesJSON)
	if err != nil {
		return PlanResult{}, err
	}
	prices, err := convertPrices(plan)
	if err != nil {
		return PlanResult{}, err
	}
	grants, err := convertGrants(limits, features)
	if err != nil {
		return PlanResult{}, err
	}
	presentation, err := offeringdomain.NewPresentation(
		plan.Description, plan.SortOrder, plan.HighlightLabel, plan.Color,
		offeringdomain.BillingInterval(plan.BillingInterval),
	)
	if err != nil {
		return PlanResult{}, ErrPlan
	}
	version, err := offeringdomain.NewVersionWithPresentation(
		plan.Name, presentation, plan.IsPublic, request.Available, request.IsBase,
		plan.TrialDays, 0, request.SnapshotAt, nil, prices, grants)
	if err != nil {
		return PlanResult{}, ErrPlan
	}
	missing := append(missingLimits, missingFeatures...)
	return PlanResult{Key: key, Version: version, MissingFields: missing}, nil
}

func convertPrices(plan PlanSnapshot) ([]offeringdomain.Price, error) {
	entry, ok := money.LookupCurrency(money.RegistryVersion, plan.Currency)
	if !ok || !entry.MinorUnitApplicable || entry.Exponent != 2 {
		return nil, ErrPlan
	}
	monthlyMoney, err := money.New(plan.PriceCents, plan.Currency, money.RegistryVersion)
	if err != nil {
		return nil, ErrPlan
	}
	monthly, err := offeringdomain.NewPrice(offeringdomain.Monthly, monthlyMoney)
	if err != nil {
		return nil, ErrPlan
	}
	prices := []offeringdomain.Price{monthly}
	if plan.AnnualPriceCents > 0 {
		annualMoney, moneyErr := money.New(
			plan.AnnualPriceCents, plan.Currency, money.RegistryVersion,
		)
		if moneyErr != nil {
			return nil, ErrPlan
		}
		annual, priceErr := offeringdomain.NewPrice(offeringdomain.Yearly, annualMoney)
		if priceErr != nil {
			return nil, ErrPlan
		}
		prices = append(prices, annual)
	}
	return prices, nil
}

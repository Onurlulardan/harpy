package billingv1

import (
	"time"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func mapPlanUpdate(
	access billingapi.Access,
	key string,
	current billingapi.OfferingView,
	input planRequest,
	now time.Time,
) (billingapi.CreateOfferingVersionCommand, error) {
	if input.Slug != nil && *input.Slug != current.Key {
		return billingapi.CreateOfferingVersionCommand{}, ErrUnsupported
	}
	name := current.Name
	if input.Name != nil {
		name = *input.Name
	}
	public := current.Public
	if input.IsPublic != nil {
		public = *input.IsPublic
	}
	available := current.Available
	if input.IsActive != nil {
		available = *input.IsActive
	}
	trial := current.TrialDays
	if input.TrialDays != nil {
		trial = *input.TrialDays
	}
	description := current.Description
	if input.Description != nil {
		description = *input.Description
	}
	sortOrder := current.SortOrder
	if input.SortOrder != nil {
		if int64(*input.SortOrder) < -1<<31 || int64(*input.SortOrder) > 1<<31-1 {
			return billingapi.CreateOfferingVersionCommand{}, ErrRequestInvalid
		}
		sortOrder = int32(*input.SortOrder)
	}
	highlight, color, interval := current.HighlightLabel, current.Color, current.BillingInterval
	if input.HighlightLabel != nil {
		highlight = *input.HighlightLabel
	}
	if input.Color != nil {
		color = *input.Color
	}
	if input.BillingInterval != nil {
		interval = *input.BillingInterval
	}
	prices, err := mergePlanPrices(current, input)
	if err != nil || !required(name) || trial < 0 {
		return billingapi.CreateOfferingVersionCommand{}, ErrRequestInvalid
	}
	entitlements, err := mergePlanEntitlements(current, input)
	if err != nil {
		return billingapi.CreateOfferingVersionCommand{}, err
	}
	return billingapi.CreateOfferingVersionCommand{
		Access: access, IdempotencyKey: key, OfferingID: current.ID,
		ExpectedRevision: current.Revision,
		Version: billingapi.OfferingVersionInput{
			Name: name, Description: description, SortOrder: sortOrder,
			HighlightLabel: highlight, Color: color, BillingInterval: interval,
			Public: public, Available: available,
			Base: current.Base, TrialDays: trial,
			GraceDays: current.GraceDays, EffectiveStart: utcMicrosecond(now),
			Prices: prices, Entitlements: entitlements,
		},
	}, nil
}

func mergePlanEntitlements(
	current billingapi.OfferingView,
	input planRequest,
) ([]billingapi.EntitlementInput, error) {
	limits, features := legacyPlanValues(current)
	if input.Limits != nil {
		decoded, _, err := decodePlanDocuments(input.Limits, nil)
		if err != nil {
			return nil, err
		}
		limits = decoded
	}
	if input.Features != nil {
		_, decoded, err := decodePlanDocuments(nil, input.Features)
		if err != nil {
			return nil, err
		}
		features = decoded
	}
	return legacyEntitlements(limits, features), nil
}

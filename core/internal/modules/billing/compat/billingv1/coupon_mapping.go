package billingv1

import (
	"context"
	"strings"
	"time"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	"github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/money"
)

func (handler *Handler) mapCouponCreate(
	ctx context.Context,
	access billingapi.Access,
	references couponReferenceAccess,
	key string,
	input couponRequest,
) (billingapi.CreateDiscountCommand, error) {
	if input.Name == nil || !required(*input.Name) || input.DiscountType == nil ||
		input.DiscountValue == nil {
		return billingapi.CreateDiscountCommand{}, ErrRequestInvalid
	}
	code, err := handler.couponCode(input.Code)
	if err != nil {
		return billingapi.CreateDiscountCommand{}, err
	}
	version, err := handler.couponVersion(
		ctx, billingapi.DiscountView{}, input, code, references,
	)
	if err != nil {
		return billingapi.CreateDiscountCommand{}, err
	}
	return billingapi.CreateDiscountCommand{
		Access: access, IdempotencyKey: key, InitialVersion: version,
	}, nil
}

func (handler *Handler) couponVersion(
	ctx context.Context,
	current billingapi.DiscountView,
	input couponRequest,
	defaultCode string,
	references couponReferenceAccess,
) (billingapi.DiscountVersionInput, error) {
	version := discountVersion(current)
	version.Code = defaultCode
	if input.Code != nil {
		version.Code = strings.ToUpper(*input.Code)
	}
	if input.Name != nil {
		version.Name = *input.Name
	}
	if input.Description != nil {
		version.Description = *input.Description
	}
	version.Duration = stringValue(input.Duration, version.Duration)
	if version.Duration == "" {
		version.Duration = "once"
	}
	if input.DurationMonths.Present {
		version.DurationCycles = input.DurationMonths.Value
	}
	if input.MaxUses.Present {
		version.GlobalUseLimit = input.MaxUses.Value
	}
	if input.MaxUsesPerTenant != nil {
		version.PerTenantUseLimit = input.MaxUsesPerTenant
	} else if current.ID == "" {
		value := int64(1)
		version.PerTenantUseLimit = &value
	}
	start := handler.clock()
	if input.ValidFrom != nil {
		start = *input.ValidFrom
	} else if current.ValidStart != "" {
		parsed, err := time.Parse(utcMicrosecondLayout, current.ValidStart)
		if err != nil {
			return billingapi.DiscountVersionInput{}, ErrRequestInvalid
		}
		start = parsed
	}
	version.ValidStart = utcMicrosecond(start)
	if input.ValidUntil.Present {
		version.ValidEnd = normalizeOptionalTime(input.ValidUntil.Value)
	}
	if input.ApplicablePlanIDs != nil {
		ids, err := handler.resolveOfferingVersions(
			ctx, references.offering, *input.ApplicablePlanIDs,
		)
		if err != nil {
			return billingapi.DiscountVersionInput{}, err
		}
		version.OfferingVersionIDs, version.AllOfferings = ids, len(ids) == 0
	}
	if current.ID == "" && input.ApplicablePlanIDs == nil {
		version.AllOfferings = true
	}
	if err := applyDiscountAmount(&version, input, current); err != nil {
		return billingapi.DiscountVersionInput{}, err
	}
	if input.IsActive != nil {
		value := *input.IsActive
		version.Available = &value
	}
	if input.CampaignID.Present {
		version.Campaign = billingapi.CampaignBindingInput{Replace: true}
		if input.CampaignID.Value != nil {
			campaignID, err := handler.canonicalLegacyID(
				ctx, references.campaign, billingapi.LegacyIdentityCampaign,
				*input.CampaignID.Value,
			)
			if err != nil {
				return billingapi.DiscountVersionInput{}, err
			}
			version.Campaign.CampaignID = &campaignID
		}
	}
	return version, nil
}

func discountVersion(current billingapi.DiscountView) billingapi.DiscountVersionInput {
	return billingapi.DiscountVersionInput{
		Name: current.Name, Description: current.Description,
		Code: current.Code, Kind: current.Kind, BasisPoints: current.BasisPoints,
		FixedAmount: current.FixedAmount, Duration: current.Duration,
		DurationCycles: current.DurationCycles, Priority: current.Priority,
		AllOfferings:       current.AllOfferings,
		OfferingVersionIDs: append([]string(nil), current.OfferingIDs...),
		ValidStart:         current.ValidStart, ValidEnd: current.ValidEnd,
		GlobalUseLimit:    current.GlobalUseLimit,
		PerTenantUseLimit: current.PerTenantUseLimit,
		Available:         boolPointer(current.Available),
		Campaign: billingapi.CampaignBindingInput{
			Replace: current.CampaignID != nil, CampaignID: current.CampaignID,
		},
	}
}

func applyDiscountAmount(
	version *billingapi.DiscountVersionInput,
	input couponRequest,
	current billingapi.DiscountView,
) error {
	kind, value := "percentage", int64(0)
	if current.BasisPoints != nil {
		value = *current.BasisPoints / 100
	} else if current.FixedAmount != nil {
		kind, value = "fixed_amount", current.FixedAmount.MinorUnits
	}
	if input.DiscountType != nil {
		kind = *input.DiscountType
	}
	if input.DiscountValue != nil {
		value = *input.DiscountValue
	}
	if value < 0 {
		return ErrRequestInvalid
	}
	version.BasisPoints, version.FixedAmount = nil, nil
	switch kind {
	case "percentage":
		if value > 100 {
			return ErrRequestInvalid
		}
		basisPoints := value * 100
		version.Kind, version.BasisPoints = "basis_points", &basisPoints
	case "fixed_amount":
		version.Kind = "fixed"
		version.FixedAmount = &billingapi.MoneyView{
			MinorUnits: value, Currency: "TRY", RegistryVersion: money.RegistryVersion,
		}
	default:
		return ErrRequestInvalid
	}
	return nil
}

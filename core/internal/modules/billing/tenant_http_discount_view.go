package billing

import (
	"math"

	discountapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/discount"
	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	"github.com/google/uuid"
)

func discountSnapshot(
	value discountapp.Eligibility,
) (billingapi.TenantDiscountSnapshot, error) {
	policy := value.Policy()
	duration := policy.Duration()
	result := billingapi.DiscountView{
		ID: value.DiscountID().String(), Code: string(policy.Code()),
		Revision: value.Revision(), VersionID: value.DiscountVersionID().String(),
		Version: value.Version(), State: "active", Duration: string(duration.Kind()),
		Priority: int(policy.Priority()), AllOfferings: policy.AllOfferings(),
		OfferingIDs:    make([]string, 0, len(policy.OfferingIDs())),
		ValidStart:     formatInstant(policy.ValidStart()),
		GlobalUseLimit: policy.GlobalLimit(), PerTenantUseLimit: policy.TenantLimit(),
	}
	if end := policy.ValidEnd(); end != nil {
		formatted := formatInstant(*end)
		result.ValidEnd = &formatted
	}
	for _, identifier := range policy.OfferingIDs() {
		result.OfferingIDs = append(result.OfferingIDs, uuid.UUID(identifier).String())
	}
	if rate, found := policy.BasisPoints(); found {
		amount := rate.Value()
		result.Kind, result.BasisPoints = "basis_points", &amount
	}
	if amount, found := policy.FixedAmount(); found {
		result.Kind = "fixed"
		result.FixedAmount = &billingapi.MoneyView{
			MinorUnits: amount.MinorUnits(), Currency: amount.Currency(),
			RegistryVersion: amount.CurrencyRegistryVersion(),
		}
	}
	if duration.Kind() == "repeating" {
		if duration.Cycles() > int64(math.MaxInt) {
			return billingapi.TenantDiscountSnapshot{}, billingapi.ErrUnavailable
		}
		cycles := int(duration.Cycles())
		result.DurationCycles = &cycles
	}
	if result.Kind == "" {
		return billingapi.TenantDiscountSnapshot{}, billingapi.ErrUnavailable
	}
	return billingapi.TenantDiscountSnapshot{
		TenantID: value.TenantID(), DiscountID: value.DiscountID(),
		DiscountVersionID: value.DiscountVersionID(),
		SubscriptionID:    value.SubscriptionID(),
		OfferingVersionID: value.OfferingVersionID(),
		PriceVersionID:    value.PriceVersionID(), Currency: value.Currency(), View: result,
	}, nil
}

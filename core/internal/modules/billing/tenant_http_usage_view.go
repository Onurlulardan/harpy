package billing

import (
	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	usageapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/usage"
)

func usageSnapshot(
	value usageapp.UsageSnapshot,
) (billingapi.TenantUsageSnapshot, error) {
	meters := value.Meters()
	if len(meters) > 512 {
		return billingapi.TenantUsageSnapshot{}, billingapi.ErrUnavailable
	}
	result := billingapi.UsageView{Meters: make([]billingapi.UsageMeterView, 0, len(meters))}
	for _, meter := range meters {
		limit := meter.Limit()
		view := billingapi.UsageMeterView{
			EntitlementKey: string(meter.Key()),
			Kind:           string(meter.Definition().ValueKind()),
			Current:        meter.Current(), Unlimited: limit.Unlimited(),
			ResetPolicy: string(meter.Definition().ResetPolicy()),
		}
		if amount, found := limit.Integer(); found {
			view.Limit = &amount
		}
		period := meter.Period()
		start, end := formatInstant(period.Start()), formatInstant(period.End())
		view.PeriodStart, view.PeriodEnd = &start, &end
		result.Meters = append(result.Meters, view)
	}
	return billingapi.TenantUsageSnapshot{TenantID: value.TenantID(), View: result}, nil
}

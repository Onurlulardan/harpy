package billing

import usageapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/usage"

type usageThresholdPayload struct {
	MeterPeriodID string `json:"meter_period_id"`
	UsageRevision int64  `json:"usage_revision"`
}

func encodeUsageThresholdEvent(
	event usageapp.ThresholdEventIntent,
) ([]byte, string, error) {
	return canonicalEventPayload(usageThresholdPayload{
		MeterPeriodID: event.MeterPeriodID.String(),
		UsageRevision: event.UsageRevision,
	})
}

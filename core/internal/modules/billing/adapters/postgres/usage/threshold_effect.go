package usage

import (
	"time"

	usageapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/usage"
	"github.com/google/uuid"
)

func validThresholdEvent(
	event *usageapp.ThresholdEventIntent,
	decision *usageapp.DecisionRecord,
	periodID uuid.UUID,
	revision int64,
	causationID uuid.UUID,
	now time.Time,
) bool {
	required := decision != nil && usageapp.ThresholdEventRequired(*decision)
	if !required {
		return event == nil
	}
	return event != nil && event.Valid() && event.MeterPeriodID == periodID &&
		event.UsageRevision == revision && event.CausationID == causationID &&
		event.OccurredAt.Equal(now)
}

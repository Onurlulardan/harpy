package usage

import (
	"time"

	"github.com/google/uuid"
)

const ThresholdReachedEventType = "billing.usage.threshold_reached"

type ThresholdEventIntent struct {
	ID            uuid.UUID
	MeterPeriodID uuid.UUID
	UsageRevision int64
	CausationID   uuid.UUID
	OccurredAt    time.Time
}

func (event ThresholdEventIntent) Valid() bool {
	return validV7(event.ID) && validV7(event.MeterPeriodID) &&
		event.UsageRevision > 0 && validV7(event.CausationID) &&
		validInstant(event.OccurredAt)
}

func (commands *Commands) thresholdEvent(
	periodID uuid.UUID,
	revision int64,
	causationID uuid.UUID,
	now time.Time,
) (ThresholdEventIntent, error) {
	eventID, err := commands.nextID()
	if err != nil || !validV7(eventID) {
		return ThresholdEventIntent{}, ErrConfiguration
	}
	event := ThresholdEventIntent{
		ID: eventID, MeterPeriodID: periodID, UsageRevision: revision,
		CausationID: causationID, OccurredAt: now,
	}
	if !event.Valid() {
		return ThresholdEventIntent{}, ErrConfiguration
	}
	return event, nil
}

func ThresholdEventRequired(decision DecisionRecord) bool {
	limit, bounded := decision.Limit.Integer()
	return bounded && decision.Allowed && decision.SoftExceeded &&
		decision.Current <= limit && decision.Requested > limit-decision.Current
}

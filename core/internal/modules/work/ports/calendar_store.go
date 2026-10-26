package ports

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
)

type CalendarStore interface {
	InsertEvent(context.Context, domain.CalendarEvent) error
	LoadEventForUpdate(context.Context, domain.ID) (domain.CalendarEvent, error)
	PersistEvent(context.Context, domain.CalendarEvent, int64) error
	InsertRecurrenceRule(context.Context, domain.ID, domain.ID, RecurrenceRevision) error
	FindRecurrenceRuleForEvent(context.Context, domain.ID) (domain.ID, int64, error)
	AppendRecurrenceRevision(context.Context, domain.ID, RecurrenceRevision, int64) error
	LoadRecurrenceRevision(context.Context, domain.ID) (RecurrenceRevision, error)
	InsertGeneratedEvents(context.Context, domain.ID, []GeneratedOccurrence) (int, error)
}

type GeneratedOccurrence struct {
	ID         domain.ID
	Occurrence domain.Occurrence
}

type RecurrenceRevision struct {
	Revision                    int64
	Contract, DTSTART, TimeZone string
	TZDBVersion, RRULE          string
	EXDATE                      []string
}

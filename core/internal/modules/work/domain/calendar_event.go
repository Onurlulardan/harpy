package domain

import (
	"strings"
	"time"
)

type CalendarEvent struct {
	ID, OwnerID, BoardID, OrganizationID ID
	Title, Description, Color            string
	Audience                             Audience
	Window                               CalendarWindow
	RecurrenceRuleID                     ID
	Generated                            bool
	RecurrenceRevision                   int64
	RowVersion                           int64
	DeletedAt                            *time.Time
}

func (event CalendarEvent) Valid() bool {
	if !event.ID.Valid() || !event.OwnerID.Valid() || strings.TrimSpace(event.Title) == "" ||
		len(event.Title) > 300 || len(event.Description) > 16_000 ||
		!ValidColor(event.Color) || !event.Audience.Valid() ||
		event.Window.Validate() != nil || event.RowVersion < 1 {
		return false
	}
	if event.Generated != (event.RecurrenceRuleID.Valid() && event.RecurrenceRevision > 0) {
		return false
	}
	switch event.Audience {
	case AudiencePrivate:
		return !event.BoardID.Valid() && !event.OrganizationID.Valid()
	case AudienceBoard:
		return event.BoardID.Valid() && !event.OrganizationID.Valid()
	case AudienceOrganization:
		return event.OrganizationID.Valid() && !event.BoardID.Valid()
	default:
		return false
	}
}

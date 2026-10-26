package postgres

import (
	"context"
	_ "embed"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
)

//go:embed sql/tenant/insert_calendar_event.sql
var insertCalendarEventSQL string

func (store *Store) InsertEvent(ctx context.Context, event domain.CalendarEvent) error {
	if !event.Valid() || event.Generated || event.DeletedAt != nil {
		return domain.ErrCalendarWindow
	}
	return executeOne(ctx, store, insertCalendarEventSQL, event.ID, event.OwnerID, event.Audience,
		nullID(event.BoardID), nullID(event.OrganizationID), event.Title, event.Description,
		event.Color, event.Window.StartsAt, event.Window.EndsAt, nullText(event.Window.StartDate),
		nullText(event.Window.EndDate), event.Window.TimeZone, event.RowVersion)
}

func nullID(id domain.ID) any {
	if !id.Valid() {
		return nil
	}
	return id
}
func nullText(value string) any {
	if value == "" {
		return nil
	}
	return value
}

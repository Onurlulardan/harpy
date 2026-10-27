package postgres

import (
	"context"
	_ "embed"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
)

//go:embed sql/tenant/update_calendar_event.sql
var updateCalendarEventSQL string

func (store *Store) PersistEvent(ctx context.Context, event domain.CalendarEvent, expected int64) error {
	if !event.Valid() || event.Generated || expected < 1 || event.RowVersion != expected+1 {
		return domain.ErrVersionConflict
	}
	return executeOne(ctx, store, updateCalendarEventSQL, event.ID, event.Audience,
		nullID(event.BoardID), nullID(event.OrganizationID), event.Title, event.Description,
		event.Color, event.Window.StartsAt, event.Window.EndsAt, nullText(event.Window.StartDate),
		nullText(event.Window.EndDate), event.Window.TimeZone, event.RowVersion, event.DeletedAt, expected)
}

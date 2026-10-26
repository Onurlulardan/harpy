package work

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/application/calendar"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func (runtime *PlatformHTTPRuntime) calendarUpdateValues(ctx context.Context,
	scope ports.Scope, eventID ports.ID, compatibility bool, wire writeWire,
) (int64, calendar.CalendarEventInput, error) {
	if !compatibility || wire.ExpectedVersion > 0 {
		value, err := calendarInput(wire, false)
		return wire.ExpectedVersion, value, err
	}
	stores, err := runtime.stores(ctx, scope)
	if err != nil {
		return 0, calendar.CalendarEventInput{}, err
	}
	event, err := stores.Calendar().LoadEventForUpdate(ctx, eventID)
	if err != nil {
		return 0, calendar.CalendarEventInput{}, err
	}
	value := calendar.CalendarEventInput{Audience: event.Audience,
		BoardID: event.BoardID, OrganizationID: event.OrganizationID,
		Title: event.Title, Description: event.Description, Color: event.Color,
		Window: event.Window}
	if wire.Title != "" {
		value.Title = wire.Title
	}
	if wire.Description != "" {
		value.Description = wire.Description
	}
	if wire.Color != "" {
		value.Color = wire.Color
	}
	if wire.StartDate != "" || wire.StartsAt != nil || wire.StartsOn != "" {
		window, windowErr := calendarWindow(wire, true)
		if windowErr != nil {
			return 0, value, ports.ErrInvalid
		}
		value.Window = window
	}
	return event.RowVersion, value, nil
}

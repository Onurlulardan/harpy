package work

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/application/calendar"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/application/recurrence"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func (runtime *PlatformHTTPRuntime) writeCalendar(ctx context.Context, scope ports.Scope,
	input HTTPWriteInput, wire writeWire,
) (ports.OperationReceipt, bool, error) {
	switch input.Operation {
	case "createWorkCalendarEvent", "createLegacyCalendarEvent":
		value, err := calendarInput(wire, input.Compatibility)
		if err != nil {
			return ports.OperationReceipt{}, true, err
		}
		receipt, err := runtime.services.calendar.CreateCalendarEvent(
			ctx, scope, mutation(input, 0), value)
		return receipt, true, err
	case "updateWorkCalendarEvent", "updateLegacyCalendarEvent":
		eventID, err := routeID(input.IDs, "event_id")
		if err != nil {
			return ports.OperationReceipt{}, true, err
		}
		version, value, err := runtime.calendarUpdateValues(
			ctx, scope, eventID, input.Compatibility, wire)
		if err != nil {
			return ports.OperationReceipt{}, true, err
		}
		receipt, err := runtime.services.calendar.UpdateCalendarEvent(
			ctx, scope, mutation(input, version), eventID, value)
		return receipt, true, err
	case "archiveLegacyCalendarEvent":
		eventID, err := routeID(input.IDs, "event_id")
		if err != nil {
			return ports.OperationReceipt{}, true, err
		}
		version, _, err := runtime.calendarUpdateValues(
			ctx, scope, eventID, true, wire)
		if err != nil {
			return ports.OperationReceipt{}, true, err
		}
		receipt, err := runtime.services.calendar.ArchiveEvent(
			ctx, scope, mutation(input, version), eventID)
		return receipt, true, err
	case "createWorkRecurrenceRule", "updateWorkRecurrenceRule":
		return runtime.writeRecurrence(ctx, scope, input, wire)
	}
	return ports.OperationReceipt{}, false, nil
}

func calendarInput(wire writeWire, compatibility bool) (calendar.CalendarEventInput, error) {
	value := calendar.CalendarEventInput{Audience: ports.Audience(wire.Audience),
		Title: wire.Title, Description: wire.Description, Color: wire.Color}
	if value.Color == "" {
		value.Color = "#64748b"
	}
	if !value.Audience.Valid() {
		value.Audience = ports.AudiencePrivate
	}
	var err error
	if wire.BoardID != "" {
		value.BoardID, err = parseID(wire.BoardID)
		value.Audience = ports.AudienceBoard
	} else if wire.OrganizationID != "" {
		value.OrganizationID, err = parseID(wire.OrganizationID)
		value.Audience = ports.AudienceOrganization
	}
	if err != nil {
		return value, err
	}
	value.Window, err = calendarWindow(wire, compatibility)
	return value, err
}

func calendarWindow(wire writeWire, compatibility bool) (ports.CalendarWindow, error) {
	zone := wire.Timezone
	if zone == "" {
		zone = "UTC"
	}
	if wire.StartsAt != nil && wire.EndsAt != nil {
		start, end := wire.StartsAt.UTC(), wire.EndsAt.UTC()
		return ports.CalendarWindow{StartsAt: &start, EndsAt: &end, TimeZone: zone}, nil
	}
	if wire.StartsOn != "" && wire.EndsOn != "" {
		return ports.CalendarWindow{StartDate: wire.StartsOn,
			EndDate: wire.EndsOn, TimeZone: zone}, nil
	}
	if compatibility && wire.StartDate != "" && wire.EndDate != "" {
		start, startErr := parseLegacyTime(wire.StartDate)
		end, endErr := parseLegacyTime(wire.EndDate)
		if startErr != nil || endErr != nil {
			return ports.CalendarWindow{}, ports.ErrInvalid
		}
		return ports.CalendarWindow{StartsAt: &start, EndsAt: &end, TimeZone: zone}, nil
	}
	return ports.CalendarWindow{}, ports.ErrInvalid
}

func parseLegacyTime(value string) (time.Time, error) {
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		parsed, err = time.Parse("2006-01-02", value)
	}
	return parsed.UTC(), err
}

func (runtime *PlatformHTTPRuntime) writeRecurrence(ctx context.Context, scope ports.Scope,
	input HTTPWriteInput, wire writeWire,
) (ports.OperationReceipt, bool, error) {
	eventID, err := parseID(wire.EventID)
	if err != nil {
		return ports.OperationReceipt{}, true, err
	}
	value := recurrence.RecurrenceInput{EventID: eventID, DTSTART: wire.DTSTART,
		TimeZone: wire.Timezone, TZDBVersion: wire.TZDBVersion,
		RRULE: wire.RRULE, EXDATE: wire.EXDATE}
	if input.Operation == "createWorkRecurrenceRule" {
		receipt, writeErr := runtime.services.recurrence.CreateRecurrence(
			ctx, scope, mutation(input, wire.ExpectedVersion), value)
		return receipt, true, writeErr
	}
	receipt, writeErr := runtime.services.recurrence.UpdateRecurrence(ctx, scope,
		mutation(input, wire.ExpectedVersion), value, wire.ExpectedRuleVersion)
	return receipt, true, writeErr
}

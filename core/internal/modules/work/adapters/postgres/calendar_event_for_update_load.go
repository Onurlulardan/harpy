package postgres

import (
	"context"
	_ "embed"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/google/uuid"
)

//go:embed sql/tenant/get_calendar_event_for_update.sql
var getCalendarEventForUpdateSQL string

func (store *Store) LoadEventForUpdate(ctx context.Context, id domain.ID) (domain.CalendarEvent, error) {
	if !id.Valid() {
		return domain.CalendarEvent{}, application.ErrInvalid
	}
	if err := store.check(ctx); err != nil {
		return domain.CalendarEvent{}, err
	}
	var eventID, ownerID uuid.UUID
	var boardID, organizationID, ruleID uuid.NullUUID
	var audience, title, description, color, zone string
	var starts, ends, startDate, endDate, deleted *time.Time
	var generated bool
	var ruleRevision *int64
	var version int64
	err := store.database.QueryRow(ctx, getCalendarEventForUpdateSQL, id).Scan(&eventID, &ownerID,
		&audience, &boardID, &organizationID, &title, &description, &color, &starts, &ends,
		&startDate, &endDate, &zone, &ruleID, &ruleRevision, &generated, &version, &deleted)
	if err != nil {
		return domain.CalendarEvent{}, err
	}
	event := domain.CalendarEvent{ID: domain.ID(eventID), OwnerID: domain.ID(ownerID),
		Audience: domain.Audience(audience), Title: title, Description: description, Color: color,
		Generated: generated, RowVersion: version}
	if boardID.Valid {
		event.BoardID = domain.ID(boardID.UUID)
	}
	if organizationID.Valid {
		event.OrganizationID = domain.ID(organizationID.UUID)
	}
	if starts != nil && ends != nil {
		start, end := starts.UTC(), ends.UTC()
		event.Window = domain.CalendarWindow{StartsAt: &start, EndsAt: &end, TimeZone: zone}
	} else {
		if startDate == nil || endDate == nil {
			return domain.CalendarEvent{}, application.ErrUnavailable
		}
		event.Window = domain.CalendarWindow{StartDate: startDate.Format("2006-01-02"), EndDate: endDate.Format("2006-01-02"), TimeZone: zone}
	}
	if ruleID.Valid {
		if ruleRevision == nil {
			return domain.CalendarEvent{}, application.ErrUnavailable
		}
		event.RecurrenceRuleID, event.RecurrenceRevision = domain.ID(ruleID.UUID), *ruleRevision
	}
	if deleted != nil {
		value := deleted.UTC()
		event.DeletedAt = &value
	}
	if !event.Valid() {
		return domain.CalendarEvent{}, application.ErrUnavailable
	}
	return event, nil
}

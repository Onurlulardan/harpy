package postgres

import (
	"context"
	_ "embed"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

//go:embed sql/tenant/load_recurrence_revision.sql
var loadRecurrenceRevisionSQL string

func (store *Store) LoadRecurrenceRevision(ctx context.Context, ruleID domain.ID) (ports.RecurrenceRevision, error) {
	if !ruleID.Valid() {
		return ports.RecurrenceRevision{}, application.ErrInvalid
	}
	if err := store.check(ctx); err != nil {
		return ports.RecurrenceRevision{}, err
	}
	var value ports.RecurrenceRevision
	err := store.database.QueryRow(ctx, loadRecurrenceRevisionSQL, ruleID).Scan(&value.Revision,
		&value.Contract, &value.DTSTART, &value.TimeZone, &value.TZDBVersion, &value.RRULE, &value.EXDATE)
	if err != nil {
		return ports.RecurrenceRevision{}, err
	}
	if !validRecurrenceRevision(value) {
		return ports.RecurrenceRevision{}, application.ErrUnavailable
	}
	return value, nil
}

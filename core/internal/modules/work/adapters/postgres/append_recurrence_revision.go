package postgres

import (
	"context"
	_ "embed"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

//go:embed sql/tenant/advance_recurrence_revision.sql
var advanceRecurrenceRevisionSQL string

func (store *Store) AppendRecurrenceRevision(ctx context.Context, ruleID domain.ID, revision ports.RecurrenceRevision, expected int64) error {
	if !ruleID.Valid() || !validRecurrenceRevision(revision) || expected < 1 || revision.Revision != expected+1 {
		return application.ErrInvalid
	}
	if err := executeOne(ctx, store, insertRecurrenceRevisionSQL, ruleID, revision.Revision,
		revision.Contract, revision.DTSTART, revision.TimeZone, revision.TZDBVersion,
		revision.RRULE, revision.EXDATE); err != nil {
		return err
	}
	return executeOne(ctx, store, advanceRecurrenceRevisionSQL, ruleID, revision.Revision, expected)
}

package postgres

import (
	"context"
	_ "embed"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

//go:embed sql/tenant/insert_recurrence_rule.sql
var insertRecurrenceRuleSQL string

//go:embed sql/tenant/insert_recurrence_revision.sql
var insertRecurrenceRevisionSQL string

func (store *Store) InsertRecurrenceRule(ctx context.Context, ruleID, eventID domain.ID, revision ports.RecurrenceRevision) error {
	if !ruleID.Valid() || !eventID.Valid() || !validRecurrenceRevision(revision) || revision.Revision != 1 {
		return application.ErrInvalid
	}
	if err := executeOne(ctx, store, insertRecurrenceRuleSQL, ruleID, eventID); err != nil {
		return err
	}
	return executeOne(ctx, store, insertRecurrenceRevisionSQL, ruleID, revision.Revision,
		revision.Contract, revision.DTSTART, revision.TimeZone, revision.TZDBVersion,
		revision.RRULE, revision.EXDATE)
}

func validRecurrenceRevision(value ports.RecurrenceRevision) bool {
	return value.Revision > 0 && value.Contract == domain.RecurrenceContract &&
		value.DTSTART != "" && value.TimeZone != "" && value.TZDBVersion != "" &&
		value.RRULE != "" && len(value.EXDATE) <= 500
}

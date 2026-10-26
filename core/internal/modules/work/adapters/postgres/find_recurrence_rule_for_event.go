package postgres

import (
	"context"
	_ "embed"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/google/uuid"
)

//go:embed sql/tenant/find_recurrence_rule_for_event.sql
var findRecurrenceRuleForEventSQL string

func (store *Store) FindRecurrenceRuleForEvent(ctx context.Context, eventID domain.ID) (domain.ID, int64, error) {
	if !eventID.Valid() {
		return domain.ID{}, 0, application.ErrInvalid
	}
	if err := store.check(ctx); err != nil {
		return domain.ID{}, 0, err
	}
	var ruleID uuid.UUID
	var revision int64
	err := store.database.QueryRow(ctx, findRecurrenceRuleForEventSQL, eventID).Scan(&ruleID, &revision)
	if err != nil || revision < 1 {
		return domain.ID{}, 0, err
	}
	return domain.ID(ruleID), revision, nil
}

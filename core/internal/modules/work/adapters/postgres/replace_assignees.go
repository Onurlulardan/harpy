package postgres

import (
	"context"
	_ "embed"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
)

//go:embed sql/tenant/deactivate_task_assignees.sql
var deactivateTaskAssigneesSQL string

//go:embed sql/tenant/insert_task_assignee.sql
var insertTaskAssigneeSQL string

func (store *Store) ReplaceAssignees(ctx context.Context, taskID domain.ID, values []domain.Principal, revision int64) error {
	if !taskID.Valid() || revision < 1 || len(values) > 100 {
		return application.ErrInvalid
	}
	if err := store.check(ctx); err != nil {
		return err
	}
	if _, err := store.database.run(ctx, deactivateTaskAssigneesSQL, taskID); err != nil {
		return err
	}
	for _, principal := range values {
		if !principal.Valid() {
			return application.ErrInvalid
		}
		if err := executeOne(ctx, store, insertTaskAssigneeSQL, taskID, principal.Kind, principal.ID); err != nil {
			return err
		}
	}
	return nil
}

package postgres

import (
	"context"
	_ "embed"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
)

//go:embed sql/tenant/persist_task_label_link.sql
var persistTaskLabelLinkSQL string

func (store *Store) PersistTaskLabelLink(ctx context.Context, taskID, labelID domain.ID, linked bool) error {
	if !taskID.Valid() || !labelID.Valid() {
		return application.ErrInvalid
	}
	if err := store.check(ctx); err != nil {
		return err
	}
	return executeOne(ctx, store, persistTaskLabelLinkSQL, taskID, labelID, linked)
}

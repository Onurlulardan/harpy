package postgres

import (
	"context"
	_ "embed"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
)

//go:embed sql/tenant/hide_inbox_projection.sql
var hideInboxProjectionSQL string

//go:embed sql/tenant/hide_calendar_projection.sql
var hideCalendarProjectionSQL string

func (store *Store) HideProjection(ctx context.Context, id domain.ID, permissionRevision int64) error {
	if !id.Valid() || permissionRevision < 1 {
		return application.ErrInvalid
	}
	if err := store.check(ctx); err != nil {
		return err
	}
	inbox, err := store.database.run(ctx, hideInboxProjectionSQL, id, permissionRevision)
	if err != nil {
		return err
	}
	calendar, err := store.database.run(ctx, hideCalendarProjectionSQL, id, permissionRevision)
	if err != nil {
		return err
	}
	if inbox.RowsAffected()+calendar.RowsAffected() != 1 {
		return application.ErrNotFound
	}
	return nil
}

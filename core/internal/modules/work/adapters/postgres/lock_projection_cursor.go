package postgres

import (
	"context"
	_ "embed"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
)

//go:embed sql/tenant/ensure_projection_cursor.sql
var ensureProjectionCursorSQL string

//go:embed sql/tenant/lock_projection_cursor.sql
var lockProjectionCursorSQL string

func (store *Store) LockCursor(ctx context.Context, owner domain.SourceOwner, kind string, id domain.ID) (domain.ProjectionCursor, error) {
	if owner == domain.SourceWork || !owner.Valid() || kind == "" || !id.Valid() {
		return domain.ProjectionCursor{}, application.ErrInvalid
	}
	if err := store.check(ctx); err != nil {
		return domain.ProjectionCursor{}, err
	}
	if _, err := store.database.run(ctx, ensureProjectionCursorSQL, owner, kind, id, make([]byte, 32)); err != nil {
		return domain.ProjectionCursor{}, err
	}
	var through int64
	if err := store.database.QueryRow(ctx, lockProjectionCursorSQL, owner, kind, id).Scan(&through); err != nil {
		return domain.ProjectionCursor{}, err
	}
	return domain.ProjectionCursor{ThroughSequence: through}, nil
}

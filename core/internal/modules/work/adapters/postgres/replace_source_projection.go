package postgres

import (
	"context"
	"crypto/sha256"
	_ "embed"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

//go:embed sql/tenant/revoke_source_inbox.sql
var revokeSourceInboxSQL string

//go:embed sql/tenant/revoke_source_calendar.sql
var revokeSourceCalendarSQL string

//go:embed sql/tenant/complete_projection_repair.sql
var completeProjectionRepairSQL string

func (store *Store) ReplaceSourceProjection(ctx context.Context, owner domain.SourceOwner, kind string,
	id domain.ID, items []ports.ProjectionItem, through int64) error {
	if owner == domain.SourceWork || !owner.Valid() || kind == "" || !id.Valid() ||
		through < 1 || len(items) > 10_000 {
		return application.ErrInvalid
	}
	if err := store.check(ctx); err != nil {
		return err
	}
	if _, err := store.database.run(ctx, revokeSourceInboxSQL, owner, kind, id); err != nil {
		return err
	}
	if _, err := store.database.run(ctx, revokeSourceCalendarSQL, owner, kind, id); err != nil {
		return err
	}
	for _, item := range items {
		if item.SourceOwner != owner || item.SourceKind != kind || item.SourceID != id || item.EventSequence > through {
			return application.ErrInvalid
		}
		if err := store.PersistProjection(ctx, item); err != nil {
			return err
		}
	}
	hash := sha256.Sum256([]byte(owner + domain.SourceOwner(":"+kind+":"+id.String())))
	return executeOne(ctx, store, completeProjectionRepairSQL, owner, kind, id, through, hash[:])
}

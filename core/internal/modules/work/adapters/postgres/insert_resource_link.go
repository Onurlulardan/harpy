package postgres

import (
	"context"
	_ "embed"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
)

//go:embed sql/tenant/insert_resource_link.sql
var insertResourceLinkSQL string

func (store *Store) InsertResourceLink(ctx context.Context, link domain.ResourceLink) error {
	if !link.Valid() {
		return domain.ErrVersionConflict
	}
	return executeOne(ctx, store, insertResourceLinkSQL, link.ID, link.WorkResourceID,
		link.TargetOwner, link.TargetKind, link.TargetID, link.TargetVersionID,
		link.ReferenceEpoch, link.PermissionRevision, link.RowVersion)
}

package postgres

import (
	"context"
	_ "embed"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/google/uuid"
)

//go:embed sql/tenant/list_task_resource_links_for_update.sql
var listTaskResourceLinksForUpdateSQL string

func (store *Store) LoadResourceLinksForUpdate(
	ctx context.Context, taskID domain.ID,
) ([]domain.ResourceLink, error) {
	if !taskID.Valid() {
		return nil, application.ErrInvalid
	}
	if err := store.check(ctx); err != nil {
		return nil, err
	}
	rows, err := store.database.Query(ctx, listTaskResourceLinksForUpdateSQL, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]domain.ResourceLink, 0, 2)
	for rows.Next() {
		var id, workID, targetID, versionID uuid.UUID
		var owner, kind string
		var epoch, permissionRevision, rowVersion int64
		if err = rows.Scan(&id, &workID, &owner, &kind, &targetID, &versionID,
			&epoch, &permissionRevision, &rowVersion); err != nil {
			return nil, err
		}
		link := domain.ResourceLink{ID: domain.ID(id), WorkResourceID: domain.ID(workID),
			TargetOwner: domain.SourceOwner(owner), TargetKind: kind,
			TargetID: domain.ID(targetID), TargetVersionID: domain.ID(versionID),
			ReferenceEpoch: epoch, PermissionRevision: permissionRevision,
			RowVersion: rowVersion}
		if !link.Valid() || link.WorkResourceID != taskID || len(result) == 2 {
			return nil, application.ErrUnavailable
		}
		result = append(result, link)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

package iam

import (
	"context"

	"github.com/google/uuid"
)

func (runtime *DocumentsArchiveAccessRuntime) RefreshDocumentsArchive(
	ctx context.Context, tenantID, actorID string, pinnedRevision uint64,
	itemIDs []string,
) error {
	tenant, err := uuid.Parse(tenantID)
	if err != nil {
		return ErrDocumentsArchiveAccess
	}
	actor, err := uuid.Parse(actorID)
	if err != nil {
		return ErrDocumentsArchiveAccess
	}
	items := make([]uuid.UUID, len(itemIDs))
	for index, itemID := range itemIDs {
		items[index], err = uuid.Parse(itemID)
		if err != nil {
			return ErrDocumentsArchiveAccess
		}
	}
	return runtime.Refresh(ctx, tenant, actor, pinnedRevision, items)
}

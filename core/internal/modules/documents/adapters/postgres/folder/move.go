package folder

import (
	"context"

	folderapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/folder"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

func (storage *store) Move(ctx context.Context, input folderapp.Mutation) (domain.Folder, bool, error) {
	if !validMutation(ctx, storage, input) || input.ExpectedVersion == 0 ||
		input.ParentID.Valid() && !input.ParentID.UUIDv7() || input.ParentID == input.FolderID {
		return domain.Folder{}, false, folderapp.ErrInvalid
	}
	return scanMutation(storage.queryRow(ctx, moveSQL,
		input.OperationID.String(), input.AuditID.String(), input.ActorID.String(),
		input.FolderID.String(), input.ExpectedVersion, nullableID(input.ParentID),
		input.IdempotencyKey, input.RequestSHA256[:], input.OccurredAt),
		storage.tenantID, storage.notFound)
}

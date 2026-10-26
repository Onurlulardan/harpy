package folder

import (
	"context"

	folderapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/folder"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

func (storage *store) Restore(ctx context.Context, input folderapp.Mutation) (domain.Folder, bool, error) {
	if !validMutation(ctx, storage, input) || input.ExpectedVersion == 0 {
		return domain.Folder{}, false, folderapp.ErrInvalid
	}
	return scanMutation(storage.queryRow(ctx, restoreSQL,
		input.OperationID.String(), input.AuditID.String(), input.ActorID.String(),
		input.FolderID.String(), input.ExpectedVersion, input.IdempotencyKey,
		input.RequestSHA256[:], input.OccurredAt), storage.tenantID, storage.notFound)
}

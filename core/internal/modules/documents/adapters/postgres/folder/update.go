package folder

import (
	"context"

	folderapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/folder"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

func (storage *store) Update(ctx context.Context, input folderapp.Mutation) (domain.Folder, bool, error) {
	if !validMutation(ctx, storage, input) || input.ExpectedVersion == 0 || input.Name == "" {
		return domain.Folder{}, false, folderapp.ErrInvalid
	}
	return scanMutation(storage.queryRow(ctx, updateSQL,
		input.OperationID.String(), input.AuditID.String(), input.ActorID.String(),
		input.FolderID.String(), input.ExpectedVersion, input.Name,
		input.IdempotencyKey, input.RequestSHA256[:], input.OccurredAt),
		storage.tenantID, storage.notFound)
}

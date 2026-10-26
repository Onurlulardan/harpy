package folder

import (
	"context"

	folderapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/folder"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

func validMutation(ctx context.Context, storage *store, input folderapp.Mutation) bool {
	return storage != nil && storage.usable(ctx) && storage.command &&
		input.ActorID == storage.actorID && input.OperationID.UUIDv7() && input.AuditID.UUIDv7() &&
		input.OperationID != input.AuditID && input.FolderID.UUIDv7() &&
		input.RequestSHA256 != [32]byte{} && input.IdempotencyKey != "" &&
		!input.OccurredAt.IsZero()
}

func nullableID(value domain.ID) any {
	if !value.Valid() {
		return nil
	}
	return value.String()
}

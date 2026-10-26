package formscleanup

import (
	"context"

	formscleanupapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/formscleanup"
)

func (storage *store) PersistRequest(ctx context.Context,
	mutation formscleanupapp.Mutation,
) (formscleanupapp.Snapshot, error) {
	if !storage.usable(ctx) || !mutation.OperationID.UUIDv7() ||
		!mutation.DefinitionID.UUIDv7() || mutation.Epoch < 2 ||
		mutation.RequestedAt.IsZero() {
		return formscleanupapp.Snapshot{}, formscleanupapp.ErrInvalid
	}
	return storage.scan(storage.queryRow(ctx, createSQL,
		mutation.OperationID.String(), mutation.DefinitionID.String(),
		mutation.Epoch, mutation.RequestedAt))
}

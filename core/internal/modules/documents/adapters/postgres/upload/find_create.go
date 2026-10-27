package upload

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	uploadapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/upload"
)

func (storage *store) FindCreate(ctx context.Context, idempotencyKey string,
	digest [32]byte,
) (uploadapp.Snapshot, bool, error) {
	if !storage.usable(ctx) || !storage.command ||
		!foundation.ValidIdempotency(idempotencyKey) || digest == [32]byte{} {
		return uploadapp.Snapshot{}, false, uploadapp.ErrInvalid
	}
	result, err := scanSnapshot(storage.queryRow(ctx, findCreateSQL,
		storage.actorID.String(), idempotencyKey, digest[:]), storage.tenantID, storage.notFound)
	if err == uploadapp.ErrNotFound {
		return uploadapp.Snapshot{}, false, nil
	}
	return result, err == nil, err
}

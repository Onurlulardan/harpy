package formscleanup

import (
	"context"
	"time"

	formscleanupapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/formscleanup"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

func (storage *store) SealReceipt(ctx context.Context, operationID domain.ID,
	now time.Time,
) (formscleanupapp.Snapshot, error) {
	if !storage.usable(ctx) || !operationID.UUIDv7() || now.IsZero() {
		return formscleanupapp.Snapshot{}, formscleanupapp.ErrInvalid
	}
	return storage.scan(storage.queryRow(ctx, finalizeSQL, operationID.String(), now))
}

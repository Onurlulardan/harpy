package upload

import (
	"context"
	"time"

	uploadapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/upload"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

func (storage *store) Cancel(ctx context.Context, uploadID, releaseReceiptID domain.ID,
	occurredAt time.Time,
) (uploadapp.CancelResult, error) {
	if !storage.usable(ctx) || !storage.command || !uploadID.UUIDv7() ||
		!releaseReceiptID.UUIDv7() || !validTime(occurredAt) {
		return uploadapp.CancelResult{}, uploadapp.ErrInvalid
	}
	stored := storedSnapshot{}
	var release bool
	fields := append(storedFields(&stored), &release)
	err := storage.queryRow(ctx, cancelSQL, uploadID.String(), storage.actorID.String(),
		releaseReceiptID.String(), occurredAt).Scan(fields...)
	if err != nil {
		if storage.notFound(err) {
			return uploadapp.CancelResult{}, uploadapp.ErrNotFound
		}
		return uploadapp.CancelResult{}, uploadapp.ErrUnavailable
	}
	snapshot, err := stored.decode(storage.tenantID)
	if err != nil {
		return uploadapp.CancelResult{}, err
	}
	result := uploadapp.CancelResult{Snapshot: snapshot, ReleaseReservation: release}
	if release {
		result.ReleaseReceiptID = releaseReceiptID
	}
	return result, nil
}

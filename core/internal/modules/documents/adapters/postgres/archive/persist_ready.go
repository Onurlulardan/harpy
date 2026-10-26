package archive

import (
	"context"

	archiveapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/archive"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

func (storage *store) PersistReady(ctx context.Context, id domain.ID,
	receipt archiveapp.ArtifactReceipt,
) (archiveapp.Snapshot, error) {
	if !storage.usable(ctx) || !storage.job || !storage.command || !id.UUIDv7() ||
		!validReceipt(receipt) {
		return archiveapp.Snapshot{}, archiveapp.ErrInvalid
	}
	rows, err := storage.execute(ctx, markReadySQL, id.String(), receipt.PlaintextBytes,
		receipt.CiphertextBytes, receipt.PlaintextSHA256[:], receipt.CiphertextSHA256[:],
		receipt.ObjectKey, receipt.ObjectVersion, int64(receipt.ObjectGeneration),
		receipt.KeyID, receipt.WrappedDEK, receipt.WrittenAt)
	if err != nil {
		return archiveapp.Snapshot{}, archiveapp.ErrUnavailable
	}
	if rows != 1 {
		result, getErr := storage.Get(ctx, id)
		if getErr == nil && result.State == domain.ArchiveReady &&
			result.ObjectVersion == receipt.ObjectVersion {
			return result, nil
		}
		return archiveapp.Snapshot{}, archiveapp.ErrConflict
	}
	return storage.Get(ctx, id)
}

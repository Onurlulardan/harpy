package archive

import (
	"context"

	archiveapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/archive"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

func (storage *store) PersistEffect(ctx context.Context, id domain.ID,
	receipt archiveapp.ArtifactReceipt,
) (archiveapp.Plan, error) {
	if !storage.usable(ctx) || !storage.job || !storage.command || !id.UUIDv7() ||
		!validReceipt(receipt) {
		return archiveapp.Plan{}, archiveapp.ErrInvalid
	}
	rows, err := storage.execute(ctx, recordEffectSQL, id.String(), receipt.PlaintextBytes,
		receipt.CiphertextBytes, receipt.PlaintextSHA256[:], receipt.CiphertextSHA256[:],
		receipt.ObjectKey, receipt.ObjectVersion, int64(receipt.ObjectGeneration),
		receipt.KeyID, receipt.WrappedDEK, receipt.WrittenAt)
	if err != nil {
		return archiveapp.Plan{}, archiveapp.ErrUnavailable
	}
	if rows != 1 {
		return archiveapp.Plan{}, archiveapp.ErrConflict
	}
	snapshot, err := storage.Get(ctx, id)
	if err != nil {
		return archiveapp.Plan{}, err
	}
	entries, err := storage.listEntries(ctx, id)
	return archiveapp.Plan{Snapshot: snapshot, Entries: entries}, err
}

func validReceipt(value archiveapp.ArtifactReceipt) bool {
	return value.ObjectKey != "" && value.ObjectVersion != "" &&
		value.ObjectGeneration > 0 && value.PlaintextBytes > 0 &&
		value.CiphertextBytes > value.PlaintextBytes &&
		value.PlaintextSHA256 != [32]byte{} && value.CiphertextSHA256 != [32]byte{} &&
		value.KeyID != "" && len(value.WrappedDEK) >= 115 && !value.WrittenAt.IsZero()
}

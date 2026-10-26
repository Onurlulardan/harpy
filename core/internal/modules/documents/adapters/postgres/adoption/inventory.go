package adoption

import (
	"context"

	adoptionapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/adoption"
)

func (storage *Store) Inventory(ctx context.Context,
	request adoptionapp.InventoryRequest,
) (adoptionapp.Checkpoint, bool, error) {
	value := request.Checkpoint
	if !storage.usable(ctx) || !storage.owns(request.Authority) ||
		value.RunID != request.Authority.RunID || !value.SourceItemID.UUIDv7() ||
		!value.SourceVersionID.UUIDv7() || !value.TargetItemID.UUIDv7() ||
		!value.TargetVersionID.UUIDv7() || value.SourceBytes < 1 ||
		value.SourceSHA256 == [32]byte{} || value.SourceMetadataSHA256 == [32]byte{} ||
		value.UpdatedAt.IsZero() {
		return adoptionapp.Checkpoint{}, false, adoptionapp.ErrInvalid
	}
	return storage.scanCheckpoint(storage.queryRow(ctx, inventorySQL,
		request.Authority.RunID.String(), int64(request.CutoverGeneration),
		request.ManifestSHA256[:], value.SourceItemID.String(),
		value.SourceVersionID.String(), value.SourceObjectKey, value.SourceRevision,
		value.SourceBytes, value.SourceSHA256[:], value.SourceMetadataSHA256[:],
		value.SourceLSN, value.TargetItemID.String(), value.TargetVersionID.String(),
		value.UpdatedAt))
}

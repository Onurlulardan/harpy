package link

import (
	"context"

	linkapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/link"
)

func (storage *store) SealFormRecordAdoption(
	ctx context.Context,
	seal linkapp.FormRecordAdoptionSeal,
) (linkapp.FormRecordAdoptionReceipt, error) {
	checkpoint := seal.Checkpoint
	if !storage.usable(ctx) || !validAdoptionCheckpoint(checkpoint, true) ||
		!checkpoint.Done || seal.SourceReceiptSHA256 == [32]byte{} ||
		seal.IssuedAt.IsZero() {
		return linkapp.FormRecordAdoptionReceipt{}, linkapp.ErrInvalid
	}
	return scanFormRecordAdoptionReceipt(storage.queryRow(ctx,
		sealFormRecordAdoptionSQL, checkpoint.OperationID.String(),
		checkpoint.DefinitionID.String(), checkpoint.Epoch,
		adoptionCursor(checkpoint.Cursor), checkpoint.RecordCount, checkpoint.LinkCount,
		checkpoint.SourceSHA256[:], checkpoint.LinkSHA256[:],
		seal.SourceReceiptSHA256[:], seal.IssuedAt))
}

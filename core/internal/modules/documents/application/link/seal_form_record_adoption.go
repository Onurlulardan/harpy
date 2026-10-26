package link

import (
	"context"

	formrecordlink "github.com/Onurlulardan/harpy/core/internal/modules/documents/api/formrecordlink"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (registry *FormRecordRegistry) SealAdoption(
	ctx context.Context,
	scope platformtx.TenantScope,
	receipt formrecordlink.SourceReceipt,
) (formrecordlink.AdoptionReceipt, error) {
	if registry == nil || !receipt.Valid() ||
		!formRecordOperation(scope.OperationID()) ||
		platformtx.RequireOwner(scope, "forms") != nil ||
		platformtx.RequireOwner(scope, "documents") != nil {
		return formrecordlink.AdoptionReceipt{}, formrecordlink.ErrInvalid
	}
	store, identity, err := registry.binder.Bind(ctx, scope)
	if err != nil || store == nil || !identity.Valid() {
		return formrecordlink.AdoptionReceipt{}, formrecordlink.ErrUnavailable
	}
	adoptionStore, ok := store.(AdoptionStore)
	if !ok {
		return formrecordlink.AdoptionReceipt{}, formrecordlink.ErrUnavailable
	}
	sealed, err := adoptionStore.SealFormRecordAdoption(ctx, FormRecordAdoptionSeal{
		Checkpoint:          adoptionCheckpoint(receipt.AdoptionProgress),
		SourceReceiptSHA256: receipt.ReceiptSHA256, IssuedAt: receipt.IssuedAt,
	})
	if err != nil || sealed.SourceReceiptSHA256 != receipt.ReceiptSHA256 ||
		sealed.ReceiptSHA256 == [32]byte{} || sealed.IssuedAt != receipt.IssuedAt {
		return formrecordlink.AdoptionReceipt{}, formrecordlink.ErrUnavailable
	}
	return formrecordlink.AdoptionReceipt{
		SourceReceiptSHA256: sealed.SourceReceiptSHA256,
		ReceiptSHA256:       sealed.ReceiptSHA256, IssuedAt: sealed.IssuedAt,
	}, nil
}

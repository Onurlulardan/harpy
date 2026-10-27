package menu

import (
	"context"
	_ "embed"

	menuapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/menu"
)

//go:embed sql/tenant/find_override_receipt.sql
var findOverrideReceiptSQL string

//go:embed sql/tenant/create_override.sql
var createOverrideSQL string

//go:embed sql/tenant/update_override.sql
var updateOverrideSQL string

//go:embed sql/tenant/record_override_receipt.sql
var recordOverrideReceiptSQL string

func (storage *store) Update(
	ctx context.Context, mutation menuapp.OverrideMutation,
) (menuapp.OverrideReceipt, error) {
	if !validMutation(mutation) {
		return menuapp.OverrideReceipt{}, menuapp.ErrInvalid
	}
	if !storage.usable(ctx, true) || storage.actorID != mutation.ActorID {
		return menuapp.OverrideReceipt{}, unavailable()
	}
	prior, found, err := storage.findReceipt(ctx, mutation)
	if err != nil {
		return menuapp.OverrideReceipt{}, err
	}
	if found {
		if prior.requestSHA256 != mutation.RequestSHA256 ||
			prior.receipt.Override.ItemID.String() != mutation.Override.ItemID.String() {
			return menuapp.OverrideReceipt{}, menuapp.ErrConflict
		}
		prior.receipt.Replayed = true
		return prior.receipt, nil
	}
	result, err := storage.applyOverride(ctx, mutation)
	if err != nil {
		return menuapp.OverrideReceipt{}, err
	}
	inserted, err := storage.recordReceipt(ctx, mutation, result)
	if err != nil || !inserted {
		return menuapp.OverrideReceipt{}, unavailable()
	}
	return result, nil
}

func (storage *store) applyOverride(
	ctx context.Context, mutation menuapp.OverrideMutation,
) (menuapp.OverrideReceipt, error) {
	statement := updateOverrideSQL
	if mutation.ExpectedRevision == 0 {
		statement = createOverrideSQL
	}
	result, err := scanOverride(storage.queryRow(ctx, statement,
		mutation.Override.ItemID.String(), mutation.Override.Label,
		mutation.Override.Order, mutation.Override.Visible, mutation.ExpectedRevision))
	if err != nil {
		if storage.notFound(err) {
			return menuapp.OverrideReceipt{}, menuapp.ErrConflict
		}
		return menuapp.OverrideReceipt{}, unavailable()
	}
	return result, nil
}

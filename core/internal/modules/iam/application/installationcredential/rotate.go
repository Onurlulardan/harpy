package installationcredential

import (
	"context"
)

func Rotate(
	ctx context.Context,
	binding RotationBinding,
	store Store,
	credential Credential,
) (RotateReceipt, error) {
	if ctx == nil || ctx.Err() != nil || nilInterface(store) || !credential.Valid() ||
		!binding.valid() {
		return RotateReceipt{}, ErrInvalid
	}
	receipt, err := store.Rotate(ctx, RotateCommand{
		Binding: binding, Credential: credential,
	})
	if err != nil {
		return RotateReceipt{}, err
	}
	if !receipt.valid(binding) {
		return RotateReceipt{}, ErrStorage
	}
	return receipt, nil
}

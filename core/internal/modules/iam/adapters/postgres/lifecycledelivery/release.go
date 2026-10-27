package lifecycledelivery

import (
	"context"

	lifecycledeliveryapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycledelivery"
)

func (store *store) Release(ctx context.Context,
	reference lifecycledeliveryapp.EffectReference) error {
	if store == nil || ctx == nil || reference.Validate() != nil {
		return lifecycledeliveryapp.ErrInvalid
	}
	result, err := store.database.Execute(ctx, releaseClaimSQL, reference.ArtifactRef,
		reference.EffectID, reference.EffectKey)
	if err != nil {
		return err
	}
	if result.RowsAffected() != 1 {
		return lifecycledeliveryapp.ErrConflict
	}
	return nil
}

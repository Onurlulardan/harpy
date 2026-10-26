package invitationdelivery

import (
	"context"

	invitationdeliveryapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/invitationdelivery"
)

func (store *store) Release(
	ctx context.Context,
	reference invitationdeliveryapp.EffectReference,
) error {
	if store == nil || ctx == nil || reference.Validate() != nil {
		return invitationdeliveryapp.ErrInvalid
	}
	result, err := store.database.Execute(
		ctx, releaseClaimSQL, reference.ArtifactRef,
		reference.EffectID, reference.EffectKey,
	)
	if err != nil {
		return err
	}
	if result.RowsAffected() != 1 {
		return invitationdeliveryapp.ErrConflict
	}
	return nil
}

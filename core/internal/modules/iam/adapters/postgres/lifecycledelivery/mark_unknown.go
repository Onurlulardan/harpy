package lifecycledelivery

import (
	"context"

	lifecycledeliveryapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycledelivery"
)

func (store *store) MarkUnknown(ctx context.Context,
	reference lifecycledeliveryapp.EffectReference) error {
	if store == nil || ctx == nil || reference.Validate() != nil {
		return lifecycledeliveryapp.ErrInvalid
	}
	result, err := store.database.Execute(ctx, markUnknownSQL, reference.ArtifactRef,
		reference.EffectID, reference.EffectKey, reference.OccurredAt)
	if err != nil {
		return err
	}
	if result.RowsAffected() != 1 {
		return lifecycledeliveryapp.ErrConflict
	}
	artifact, err := artifactID(reference.ArtifactRef)
	if err != nil {
		return lifecycledeliveryapp.ErrInvalid
	}
	return store.recordSetupOutcome(ctx, reference.TenantID, artifact,
		"outcome_unknown")
}

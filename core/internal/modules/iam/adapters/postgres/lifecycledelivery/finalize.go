package lifecycledelivery

import (
	"context"

	lifecycledeliveryapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycledelivery"
)

func (store *store) Finalize(ctx context.Context,
	command lifecycledeliveryapp.FinalizeCommand,
) (lifecycledeliveryapp.StoredResult, error) {
	if store == nil || ctx == nil || command.Validate() != nil {
		return lifecycledeliveryapp.StoredResult{}, lifecycledeliveryapp.ErrInvalid
	}
	if replay, err := store.readReceipt(ctx, lifecycledeliveryapp.ClaimCommand{
		TenantID: command.TenantID, ArtifactRef: command.ArtifactRef,
		EffectID: command.EffectID, EffectKey: command.EffectKey,
		MessageID: command.Result.MessageID, ClaimedAt: command.OccurredAt,
	}); err == nil {
		return replay, nil
	} else if !rowNotFound(err) {
		return lifecycledeliveryapp.StoredResult{}, err
	}
	artifact, err := artifactID(command.ArtifactRef)
	if err != nil {
		return lifecycledeliveryapp.StoredResult{}, lifecycledeliveryapp.ErrInvalid
	}
	result, err := store.database.Execute(ctx, insertReceiptSQL,
		command.Result.ReceiptID, artifact, command.ArtifactRef,
		command.EffectID, command.EffectKey, command.Result.MessageID,
		command.Result.Outcome, command.Result.Code,
		command.Result.ProviderReference, command.Result.ResultSHA256[:],
		command.Result.CompletedAt)
	if err != nil {
		return lifecycledeliveryapp.StoredResult{}, err
	}
	if result.RowsAffected() != 1 {
		return lifecycledeliveryapp.StoredResult{}, lifecycledeliveryapp.ErrConflict
	}
	deleted, err := store.database.Execute(ctx, deleteArtifactSQL,
		artifact, command.ArtifactRef, command.EffectID, command.EffectKey)
	if err != nil {
		return lifecycledeliveryapp.StoredResult{}, err
	}
	if deleted.RowsAffected() != 1 {
		return lifecycledeliveryapp.StoredResult{}, lifecycledeliveryapp.ErrConflict
	}
	if err := store.recordSetupOutcome(ctx, command.TenantID, artifact,
		string(command.Result.Outcome)); err != nil {
		return lifecycledeliveryapp.StoredResult{}, err
	}
	return command.Result, nil
}

package invitationdelivery

import (
	"context"
	invitationdeliveryapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/invitationdelivery"
)

func (store *store) Finalize(
	ctx context.Context,
	command invitationdeliveryapp.FinalizeCommand,
) (invitationdeliveryapp.StoredResult, error) {
	if store == nil || ctx == nil || command.Validate() != nil {
		return invitationdeliveryapp.StoredResult{}, invitationdeliveryapp.ErrInvalid
	}
	if replay, err := store.readReceipt(ctx, invitationdeliveryapp.ClaimCommand{
		TenantID: command.TenantID, ArtifactRef: command.ArtifactRef,
		EffectID: command.EffectID, EffectKey: command.EffectKey,
		MessageID: command.Result.MessageID, ClaimedAt: command.OccurredAt,
	}); err == nil {
		return replay, nil
	} else if !rowNotFound(err) {
		return invitationdeliveryapp.StoredResult{}, err
	}
	result, err := store.scanReceipt(
		ctx, finalizeSQL, command.Result.ReceiptID,
		command.ArtifactRef, command.EffectID, command.EffectKey,
		command.Result.MessageID, command.Result.Outcome, command.Result.Code,
		command.Result.ProviderReference, command.Result.ResultSHA256[:],
		command.Result.CompletedAt,
	)
	if err != nil {
		return invitationdeliveryapp.StoredResult{}, err
	}
	if result.Validate(command.TenantID) != nil ||
		result.ArtifactRef != command.ArtifactRef ||
		result.EffectKey != command.EffectKey ||
		result.MessageID != command.Result.MessageID {
		return invitationdeliveryapp.StoredResult{}, invitationdeliveryapp.ErrConflict
	}
	return result, nil
}

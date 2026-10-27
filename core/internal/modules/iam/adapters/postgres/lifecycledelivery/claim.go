package lifecycledelivery

import (
	"context"

	lifecycledeliveryapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycledelivery"
)

func (store *store) Claim(ctx context.Context,
	command lifecycledeliveryapp.ClaimCommand,
) (lifecycledeliveryapp.ClaimDecision, error) {
	if store == nil || ctx == nil || command.Validate() != nil {
		return lifecycledeliveryapp.ClaimDecision{}, lifecycledeliveryapp.ErrInvalid
	}
	if replay, err := store.readReceipt(ctx, command); err == nil {
		return lifecycledeliveryapp.ClaimDecision{Replay: &replay}, nil
	} else if !rowNotFound(err) {
		return lifecycledeliveryapp.ClaimDecision{}, err
	}
	artifact, err := store.readArtifact(ctx, command)
	if rowNotFound(err) {
		return lifecycledeliveryapp.ClaimDecision{},
			lifecycledeliveryapp.ErrArtifactUnavailable
	}
	if err != nil {
		return lifecycledeliveryapp.ClaimDecision{}, err
	}
	result, err := store.database.Execute(ctx, claimArtifactSQL,
		artifact.Envelope.ArtifactID, command.ArtifactRef, command.EffectID,
		command.EffectKey, command.MessageID, command.ClaimedAt)
	if err != nil {
		return lifecycledeliveryapp.ClaimDecision{}, err
	}
	if result.RowsAffected() == 1 {
		return lifecycledeliveryapp.ClaimDecision{Artifact: &artifact}, nil
	}
	if err := store.assertHeldClaim(ctx, command); err != nil {
		return lifecycledeliveryapp.ClaimDecision{}, err
	}
	return lifecycledeliveryapp.ClaimDecision{Held: true}, nil
}

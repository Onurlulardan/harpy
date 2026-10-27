package invitationdelivery

import (
	"context"
	invitationdeliveryapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/invitationdelivery"
)

func (store *store) Claim(
	ctx context.Context,
	command invitationdeliveryapp.ClaimCommand,
) (invitationdeliveryapp.ClaimDecision, error) {
	if store == nil || ctx == nil || command.Validate() != nil {
		return invitationdeliveryapp.ClaimDecision{}, invitationdeliveryapp.ErrInvalid
	}
	if replay, err := store.readReceipt(ctx, command); err == nil {
		return invitationdeliveryapp.ClaimDecision{Replay: &replay}, nil
	} else if !rowNotFound(err) {
		return invitationdeliveryapp.ClaimDecision{}, err
	}
	artifact, err := store.claimArtifact(ctx, command)
	if rowNotFound(err) {
		if heldErr := store.assertHeldClaim(ctx, command); heldErr == nil {
			return invitationdeliveryapp.ClaimDecision{Held: true}, nil
		} else if rowNotFound(heldErr) {
			return invitationdeliveryapp.ClaimDecision{}, invitationdeliveryapp.ErrArtifactUnavailable
		} else {
			return invitationdeliveryapp.ClaimDecision{}, heldErr
		}
	}
	if err != nil {
		return invitationdeliveryapp.ClaimDecision{}, err
	}
	return invitationdeliveryapp.ClaimDecision{Artifact: &artifact}, nil
}

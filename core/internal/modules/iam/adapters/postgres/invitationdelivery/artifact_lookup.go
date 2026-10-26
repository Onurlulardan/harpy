package invitationdelivery

import (
	"context"

	deliveryartifactv1 "github.com/Onurlulardan/harpy/contracts/deliveryartifact/v1"
	invitationdeliveryapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/invitationdelivery"
)

func (store *store) claimArtifact(
	ctx context.Context,
	command invitationdeliveryapp.ClaimCommand,
) (invitationdeliveryapp.Artifact, error) {
	var artifact invitationdeliveryapp.Artifact
	var nonce, ciphertext, tag, aad, sealedDigest []byte
	err := store.database.QueryRow(
		ctx, claimArtifactSQL, command.ArtifactRef, command.EffectID,
		command.EffectKey, command.MessageID, command.ClaimedAt,
	).Scan(
		&artifact.Envelope.ArtifactID, &artifact.Envelope.InvitationID,
		&artifact.Envelope.TokenID, &artifact.Reference, &artifact.Recipient,
		&artifact.Envelope.KeyID, &nonce, &ciphertext, &tag, &aad, &sealedDigest,
		&artifact.ExpiresAt,
	)
	if err != nil {
		return invitationdeliveryapp.Artifact{}, err
	}
	if len(nonce) != deliveryartifactv1.NonceBytes ||
		len(ciphertext) != deliveryartifactv1.CiphertextBytes ||
		len(tag) != deliveryartifactv1.TagBytes || len(aad) != 32 ||
		len(sealedDigest) != 32 {
		return invitationdeliveryapp.Artifact{}, invitationdeliveryapp.ErrInvalid
	}
	copy(artifact.Envelope.Nonce[:], nonce)
	copy(artifact.Envelope.Ciphertext[:], ciphertext)
	copy(artifact.Envelope.Tag[:], tag)
	copy(artifact.Envelope.AADSHA256[:], aad)
	copy(artifact.Envelope.CiphertextSHA256[:], sealedDigest)
	if artifact.Validate(command.TenantID) != nil {
		return invitationdeliveryapp.Artifact{}, invitationdeliveryapp.ErrInvalid
	}
	return artifact, nil
}

func (store *store) assertHeldClaim(
	ctx context.Context,
	command invitationdeliveryapp.ClaimCommand,
) error {
	var artifactRef string
	var effectID string
	var effectKey string
	var messageID string
	var state string
	err := store.database.QueryRow(ctx, readClaimSQL, command.ArtifactRef).Scan(
		&artifactRef, &effectID, &effectKey, &messageID, &state,
	)
	if err != nil {
		return err
	}
	if artifactRef != command.ArtifactRef || effectKey == "" ||
		messageID == "" || (state != "claimed" && state != "outcome_unknown") ||
		effectID == "" {
		return invitationdeliveryapp.ErrConflict
	}
	return nil
}

package lifecycledelivery

import (
	"context"

	lifecycledeliveryv1 "github.com/Onurlulardan/harpy/contracts/lifecycledelivery/v1"
	lifecycledeliveryapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycledelivery"
)

func (store *store) readArtifact(ctx context.Context,
	command lifecycledeliveryapp.ClaimCommand,
) (lifecycledeliveryapp.Artifact, error) {
	var artifact lifecycledeliveryapp.Artifact
	var purpose string
	var nonce, ciphertext, tag, aad, sealedDigest []byte
	err := store.database.QueryRow(ctx, readArtifactSQL, command.ArtifactRef).Scan(
		&artifact.Envelope.ArtifactID, &artifact.Envelope.SubjectUserID,
		&artifact.Envelope.ResourceID, &artifact.Envelope.TokenID, &purpose,
		&artifact.Reference, &artifact.Recipient, &artifact.Envelope.KeyID,
		&nonce, &ciphertext, &tag, &aad, &sealedDigest, &artifact.ExpiresAt,
	)
	defer clear(nonce)
	defer clear(ciphertext)
	defer clear(tag)
	defer clear(aad)
	defer clear(sealedDigest)
	if err != nil {
		return lifecycledeliveryapp.Artifact{}, err
	}
	if len(nonce) != lifecycledeliveryv1.NonceBytes ||
		len(ciphertext) != lifecycledeliveryv1.CiphertextBytes ||
		len(tag) != lifecycledeliveryv1.TagBytes || len(aad) != 32 ||
		len(sealedDigest) != 32 {
		return lifecycledeliveryapp.Artifact{}, lifecycledeliveryapp.ErrInvalid
	}
	artifact.Envelope.Purpose = lifecycledeliveryv1.Purpose(purpose)
	copy(artifact.Envelope.Nonce[:], nonce)
	copy(artifact.Envelope.Ciphertext[:], ciphertext)
	copy(artifact.Envelope.Tag[:], tag)
	copy(artifact.Envelope.AADSHA256[:], aad)
	copy(artifact.Envelope.CiphertextSHA256[:], sealedDigest)
	if artifact.Validate(command.TenantID) != nil {
		return lifecycledeliveryapp.Artifact{}, lifecycledeliveryapp.ErrInvalid
	}
	return artifact, nil
}

func (store *store) assertHeldClaim(ctx context.Context,
	command lifecycledeliveryapp.ClaimCommand) error {
	var artifactRef, effectID, effectKey, messageID, state string
	err := store.database.QueryRow(ctx, readClaimSQL, command.ArtifactRef).Scan(
		&artifactRef, &effectID, &effectKey, &messageID, &state,
	)
	if err != nil {
		return err
	}
	if artifactRef != command.ArtifactRef || effectID == "" || effectKey == "" ||
		messageID == "" || (state != "claimed" && state != "outcome_unknown") {
		return lifecycledeliveryapp.ErrConflict
	}
	return nil
}

package gcpkms

import (
	"bytes"
	"context"
	"io"

	"cloud.google.com/go/kms/apiv1/kmspb"
	lifecycledeliveryv1 "github.com/Onurlulardan/harpy/contracts/lifecycledelivery/v1"
	"github.com/google/uuid"
)

func (sealer *DeliverySealer) SealLifecycleCredential(
	ctx context.Context,
	tenantID, artifactID, subjectUserID, resourceID, tokenID uuid.UUID,
	purpose lifecycledeliveryv1.Purpose,
	plaintext []byte,
) (lifecycledeliveryv1.Envelope, error) {
	if sealer == nil || sealer.client == nil || sealer.closed.Load() ||
		ctx == nil || ctx.Err() != nil || len(plaintext) != lifecycledeliveryv1.PlaintextBytes {
		return lifecycledeliveryv1.Envelope{}, ErrContract
	}
	keyID, err := sealer.active.Value()
	if err != nil {
		return lifecycledeliveryv1.Envelope{}, ErrContract
	}
	resource, err := sealer.active.Resource()
	if err != nil {
		return lifecycledeliveryv1.Envelope{}, ErrContract
	}
	aad, err := lifecycledeliveryv1.AADSHA256(
		tenantID, artifactID, subjectUserID, resourceID, tokenID, purpose, keyID,
	)
	if err != nil {
		return lifecycledeliveryv1.Envelope{}, ErrContract
	}
	iv := make([]byte, lifecycledeliveryv1.NonceBytes)
	if _, err = io.ReadFull(sealer.randomness, iv); err != nil ||
		bytes.Equal(iv, make([]byte, len(iv))) {
		clear(iv)
		return lifecycledeliveryv1.Envelope{}, ErrUnavailable
	}
	defer clear(iv)
	secret := append([]byte(nil), plaintext...)
	defer clear(secret)
	response, err := sealer.client.RawEncrypt(ctx, &kmspb.RawEncryptRequest{
		Name: resource, Plaintext: secret, AdditionalAuthenticatedData: aad[:],
		InitializationVector: iv, PlaintextCrc32C: checksum(secret),
		AdditionalAuthenticatedDataCrc32C: checksum(aad[:]),
		InitializationVectorCrc32C:        checksum(iv),
	})
	if err != nil {
		return lifecycledeliveryv1.Envelope{}, ErrUnavailable
	}
	return verifiedLifecycleEnvelope(response, tenantID, artifactID, subjectUserID,
		resourceID, tokenID, purpose, keyID, resource, aad, iv)
}

func verifiedLifecycleEnvelope(response *kmspb.RawEncryptResponse,
	tenantID, artifactID, subjectUserID, resourceID, tokenID uuid.UUID,
	purpose lifecycledeliveryv1.Purpose, keyID, resource string,
	aad [32]byte, requestedIV []byte) (lifecycledeliveryv1.Envelope, error) {
	if response == nil {
		return lifecycledeliveryv1.Envelope{}, ErrIntegrity
	}
	defer clear(response.Ciphertext)
	defer clear(response.InitializationVector)
	if !response.VerifiedPlaintextCrc32C ||
		!response.VerifiedAdditionalAuthenticatedDataCrc32C ||
		!response.VerifiedInitializationVectorCrc32C || response.Name != resource ||
		response.TagLength != lifecycledeliveryv1.TagBytes ||
		len(response.Ciphertext) != lifecycledeliveryv1.CiphertextBytes+lifecycledeliveryv1.TagBytes ||
		len(response.InitializationVector) != lifecycledeliveryv1.NonceBytes ||
		!bytes.Equal(response.InitializationVector, requestedIV) ||
		!checksumMatches(response.InitializationVector, response.InitializationVectorCrc32C) ||
		!checksumMatches(response.Ciphertext, response.CiphertextCrc32C) {
		return lifecycledeliveryv1.Envelope{}, ErrIntegrity
	}
	var nonce [lifecycledeliveryv1.NonceBytes]byte
	var ciphertext [lifecycledeliveryv1.CiphertextBytes]byte
	var tag [lifecycledeliveryv1.TagBytes]byte
	copy(nonce[:], response.InitializationVector)
	copy(ciphertext[:], response.Ciphertext[:lifecycledeliveryv1.CiphertextBytes])
	copy(tag[:], response.Ciphertext[lifecycledeliveryv1.CiphertextBytes:])
	envelope := lifecycledeliveryv1.NewEnvelope(artifactID, subjectUserID,
		resourceID, tokenID, purpose, keyID, nonce, ciphertext, tag, aad)
	if envelope.Validate(tenantID) != nil {
		return lifecycledeliveryv1.Envelope{}, ErrIntegrity
	}
	return envelope, nil
}

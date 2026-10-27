package gcpkms

import (
	"bytes"
	"context"
	"io"

	"cloud.google.com/go/kms/apiv1/kmspb"
	deliveryartifactv1 "github.com/Onurlulardan/harpy/contracts/deliveryartifact/v1"
	"github.com/google/uuid"
)

func (sealer *DeliverySealer) SealInvitationCredential(
	ctx context.Context,
	tenantID, artifactID, invitationID, tokenID uuid.UUID,
	plaintext []byte,
) (deliveryartifactv1.Envelope, error) {
	if sealer == nil || sealer.client == nil || sealer.closed.Load() ||
		ctx == nil || ctx.Err() != nil || tenantID == uuid.Nil ||
		artifactID == uuid.Nil || invitationID == uuid.Nil || tokenID == uuid.Nil ||
		len(plaintext) != deliveryartifactv1.PlaintextBytes {
		return deliveryartifactv1.Envelope{}, ErrContract
	}
	keyID, err := sealer.active.Value()
	if err != nil {
		return deliveryartifactv1.Envelope{}, ErrContract
	}
	resource, err := sealer.active.Resource()
	if err != nil {
		return deliveryartifactv1.Envelope{}, ErrContract
	}
	aad, err := deliveryartifactv1.AADSHA256(
		tenantID, artifactID, invitationID, tokenID, keyID,
	)
	if err != nil {
		return deliveryartifactv1.Envelope{}, ErrContract
	}
	iv := make([]byte, deliveryartifactv1.NonceBytes)
	if _, err = io.ReadFull(sealer.randomness, iv); err != nil ||
		bytes.Equal(iv, make([]byte, len(iv))) {
		clear(iv)
		return deliveryartifactv1.Envelope{}, ErrUnavailable
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
		return deliveryartifactv1.Envelope{}, ErrUnavailable
	}
	return verifiedDeliveryEnvelope(
		response, tenantID, artifactID, invitationID, tokenID, keyID, resource, aad, iv,
	)
}

func verifiedDeliveryEnvelope(
	response *kmspb.RawEncryptResponse,
	tenantID, artifactID, invitationID, tokenID uuid.UUID,
	keyID, resource string,
	aad [32]byte,
	requestedIV []byte,
) (deliveryartifactv1.Envelope, error) {
	if response == nil {
		return deliveryartifactv1.Envelope{}, ErrIntegrity
	}
	defer clear(response.Ciphertext)
	defer clear(response.InitializationVector)
	if !response.VerifiedPlaintextCrc32C ||
		!response.VerifiedAdditionalAuthenticatedDataCrc32C ||
		!response.VerifiedInitializationVectorCrc32C || response.Name != resource ||
		response.TagLength != deliveryartifactv1.TagBytes ||
		len(response.Ciphertext) != deliveryartifactv1.CiphertextBytes+deliveryartifactv1.TagBytes ||
		len(response.InitializationVector) != deliveryartifactv1.NonceBytes ||
		!bytes.Equal(response.InitializationVector, requestedIV) ||
		!checksumMatches(response.Ciphertext, response.CiphertextCrc32C) ||
		!checksumMatches(response.InitializationVector, response.InitializationVectorCrc32C) {
		return deliveryartifactv1.Envelope{}, ErrIntegrity
	}
	var nonce [deliveryartifactv1.NonceBytes]byte
	var ciphertext [deliveryartifactv1.CiphertextBytes]byte
	var tag [deliveryartifactv1.TagBytes]byte
	copy(nonce[:], response.InitializationVector)
	copy(ciphertext[:], response.Ciphertext[:deliveryartifactv1.CiphertextBytes])
	copy(tag[:], response.Ciphertext[deliveryartifactv1.CiphertextBytes:])
	envelope := deliveryartifactv1.NewEnvelope(
		artifactID, invitationID, tokenID, keyID, nonce, ciphertext, tag, aad,
	)
	if envelope.Validate(tenantID) != nil {
		return deliveryartifactv1.Envelope{}, ErrIntegrity
	}
	return envelope, nil
}

package gcpkms

import (
	"bytes"
	"context"
	"io"

	"cloud.google.com/go/kms/apiv1/kmspb"
	"github.com/google/uuid"
)

type TOTPEnvelope struct {
	KMSKeyID          string
	Nonce             [totpNonceBytes]byte
	Ciphertext        [totpCiphertextBytes]byte
	AuthenticationTag [totpTagBytes]byte
	AADSHA256         [32]byte
}

func (sealer *TOTPSealer) SealTOTP(
	ctx context.Context,
	environmentID, tenantID, userID, authenticatorID uuid.UUID,
	plaintext [totpPlaintextBytes]byte,
) (TOTPEnvelope, error) {
	if sealer == nil || sealer.client == nil || sealer.closed.Load() ||
		ctx == nil || ctx.Err() != nil || plaintext == [totpPlaintextBytes]byte{} {
		return TOTPEnvelope{}, ErrContract
	}
	keyID, err := sealer.active.Value()
	if err != nil {
		return TOTPEnvelope{}, ErrContract
	}
	resource, err := sealer.active.Resource()
	if err != nil {
		return TOTPEnvelope{}, ErrContract
	}
	aad, err := TOTPAdditionalAuthenticatedData(
		environmentID, tenantID, userID, authenticatorID, keyID,
	)
	if err != nil {
		return TOTPEnvelope{}, err
	}
	iv := make([]byte, totpNonceBytes)
	if _, err = io.ReadFull(sealer.randomness, iv); err != nil ||
		bytes.Equal(iv, make([]byte, len(iv))) {
		clear(iv)
		return TOTPEnvelope{}, ErrUnavailable
	}
	defer clear(iv)
	secret := append([]byte(nil), plaintext[:]...)
	defer clear(secret)
	response, err := sealer.client.RawEncrypt(ctx, &kmspb.RawEncryptRequest{
		Name: resource, Plaintext: secret, AdditionalAuthenticatedData: aad[:],
		InitializationVector: iv, PlaintextCrc32C: checksum(secret),
		AdditionalAuthenticatedDataCrc32C: checksum(aad[:]),
		InitializationVectorCrc32C:        checksum(iv),
	})
	if err != nil {
		return TOTPEnvelope{}, ErrUnavailable
	}
	return verifyTOTPEnvelope(response, keyID, resource, aad, iv)
}

func verifyTOTPEnvelope(
	response *kmspb.RawEncryptResponse,
	keyID, resource string,
	aad [32]byte,
	requestedIV []byte,
) (TOTPEnvelope, error) {
	if response == nil {
		return TOTPEnvelope{}, ErrIntegrity
	}
	defer clear(response.Ciphertext)
	defer clear(response.InitializationVector)
	if !response.VerifiedPlaintextCrc32C ||
		!response.VerifiedAdditionalAuthenticatedDataCrc32C ||
		!response.VerifiedInitializationVectorCrc32C || response.Name != resource ||
		response.TagLength != totpTagBytes ||
		len(response.Ciphertext) != totpCiphertextBytes+totpTagBytes ||
		len(response.InitializationVector) != totpNonceBytes ||
		!bytes.Equal(response.InitializationVector, requestedIV) ||
		!checksumMatches(response.InitializationVector, response.InitializationVectorCrc32C) ||
		!checksumMatches(response.Ciphertext, response.CiphertextCrc32C) {
		return TOTPEnvelope{}, ErrIntegrity
	}
	result := TOTPEnvelope{KMSKeyID: keyID, AADSHA256: aad}
	copy(result.Nonce[:], response.InitializationVector)
	copy(result.Ciphertext[:], response.Ciphertext[:totpCiphertextBytes])
	copy(result.AuthenticationTag[:], response.Ciphertext[totpCiphertextBytes:])
	return result, nil
}

package gcpkms

import (
	"context"

	"cloud.google.com/go/kms/apiv1/kmspb"
)

const (
	totpNonceBytes      = 12
	totpCiphertextBytes = 32
	totpTagBytes        = 16
	totpPlaintextBytes  = 32
)

func (decryptor *TOTPDecryptor) DecryptTOTP(
	ctx context.Context,
	keyID string,
	nonce []byte,
	ciphertext []byte,
	tag []byte,
	aadSHA256 [32]byte,
) ([32]byte, error) {
	if decryptor == nil || decryptor.client == nil || decryptor.closed.Load() ||
		ctx == nil || ctx.Err() != nil || len(nonce) != totpNonceBytes ||
		len(ciphertext) != totpCiphertextBytes || len(tag) != totpTagBytes ||
		aadSHA256 == [32]byte{} {
		return [32]byte{}, ErrContract
	}
	key, err := parseKeyVersion(keyID)
	if err != nil || !decryptor.active.SameCryptoKey(key) {
		return [32]byte{}, ErrContract
	}
	resource, err := key.Resource()
	if err != nil {
		return [32]byte{}, ErrContract
	}
	sealed := make([]byte, 0, len(ciphertext)+len(tag))
	sealed = append(sealed, ciphertext...)
	sealed = append(sealed, tag...)
	defer clear(sealed)
	request := &kmspb.RawDecryptRequest{
		Name: resource, Ciphertext: sealed,
		AdditionalAuthenticatedData: aadSHA256[:],
		InitializationVector:        nonce, TagLength: totpTagBytes,
		CiphertextCrc32C:                  checksum(sealed),
		AdditionalAuthenticatedDataCrc32C: checksum(aadSHA256[:]),
		InitializationVectorCrc32C:        checksum(nonce),
	}
	response, err := decryptor.client.RawDecrypt(ctx, request)
	if err != nil {
		return [32]byte{}, ErrUnavailable
	}
	return verifiedPlaintext(response)
}

func verifiedPlaintext(response *kmspb.RawDecryptResponse) ([32]byte, error) {
	if response == nil {
		return [32]byte{}, ErrIntegrity
	}
	defer clear(response.Plaintext)
	if !response.VerifiedCiphertextCrc32C ||
		!response.VerifiedAdditionalAuthenticatedDataCrc32C ||
		!response.VerifiedInitializationVectorCrc32C ||
		len(response.Plaintext) != totpPlaintextBytes ||
		!checksumMatches(response.Plaintext, response.PlaintextCrc32C) {
		return [32]byte{}, ErrIntegrity
	}
	var plaintext [32]byte
	copy(plaintext[:], response.Plaintext)
	return plaintext, nil
}

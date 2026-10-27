package main

import (
	"bytes"
	"context"
	"testing"

	"cloud.google.com/go/kms/apiv1/kmspb"
)

const fixtureResource = "projects/harpy-local/locations/europe-west1/" +
	"keyRings/iam/cryptoKeys/runtime/cryptoKeyVersions/1"

var (
	fixtureKey       = bytes.Repeat([]byte{0x31}, aesKeyBytes)
	fixturePlaintext = bytes.Repeat([]byte{0x42}, 32)
	fixtureAAD       = bytes.Repeat([]byte{0x53}, 32)
	fixtureIV        = bytes.Repeat([]byte{0x64}, gcmNonceBytes)
)

func validEncryptFixture() *kmspb.RawEncryptRequest {
	return &kmspb.RawEncryptRequest{
		Name:                              fixtureResource,
		Plaintext:                         append([]byte(nil), fixturePlaintext...),
		AdditionalAuthenticatedData:       append([]byte(nil), fixtureAAD...),
		InitializationVector:              append([]byte(nil), fixtureIV...),
		PlaintextCrc32C:                   checksum(fixturePlaintext),
		AdditionalAuthenticatedDataCrc32C: checksum(fixtureAAD),
		InitializationVectorCrc32C:        checksum(fixtureIV),
	}
}

func encryptedFixture(t *testing.T) []byte {
	t.Helper()
	response, err := newKMSService(fixtureKey).RawEncrypt(
		context.Background(), validEncryptFixture(),
	)
	if err != nil {
		t.Fatalf("encrypt fixture: %v", err)
	}
	return response.Ciphertext
}

func validDecryptFixture(t *testing.T) *kmspb.RawDecryptRequest {
	t.Helper()
	ciphertext := encryptedFixture(t)
	return &kmspb.RawDecryptRequest{
		Name:                              fixtureResource,
		Ciphertext:                        ciphertext,
		AdditionalAuthenticatedData:       append([]byte(nil), fixtureAAD...),
		InitializationVector:              append([]byte(nil), fixtureIV...),
		TagLength:                         gcmTagBytes,
		CiphertextCrc32C:                  checksum(ciphertext),
		AdditionalAuthenticatedDataCrc32C: checksum(fixtureAAD),
		InitializationVectorCrc32C:        checksum(fixtureIV),
	}
}

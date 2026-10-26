package gcpkms

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"cloud.google.com/go/kms/apiv1/kmspb"
	"github.com/googleapis/gax-go/v2"
	"google.golang.org/protobuf/proto"
)

type decryptClientFixture struct {
	request  *kmspb.RawDecryptRequest
	response *kmspb.RawDecryptResponse
	err      error
	closed   int
}

func (client *decryptClientFixture) RawDecrypt(
	_ context.Context,
	request *kmspb.RawDecryptRequest,
	_ ...gax.CallOption,
) (*kmspb.RawDecryptResponse, error) {
	client.request = proto.Clone(request).(*kmspb.RawDecryptRequest)
	return client.response, client.err
}

func (client *decryptClientFixture) Close() error {
	client.closed++
	return nil
}

func TestDecryptTOTPUsesRawAESGCMAndVerifiesIntegrity(t *testing.T) {
	plaintext := bytes.Repeat([]byte{0x41}, totpPlaintextBytes)
	client := &decryptClientFixture{response: validResponse(plaintext)}
	decryptor := newFixtureDecryptor(t, client)
	nonce := bytes.Repeat([]byte{0x11}, totpNonceBytes)
	ciphertext := bytes.Repeat([]byte{0x22}, totpCiphertextBytes)
	tag := bytes.Repeat([]byte{0x33}, totpTagBytes)
	aad := [32]byte{0x44}
	result, err := decryptor.DecryptTOTP(
		context.Background(), retiredFixtureKey, nonce, ciphertext, tag, aad,
	)
	if err != nil {
		t.Fatalf("decrypt TOTP: %v", err)
	}
	if !bytes.Equal(result[:], plaintext) {
		t.Fatal("unexpected plaintext")
	}
	request := client.request
	if request == nil || request.Name == retiredFixtureKey || request.TagLength != 16 ||
		!bytes.Equal(request.Ciphertext, append(bytes.Clone(ciphertext), tag...)) ||
		!checksumMatches(request.Ciphertext, request.CiphertextCrc32C) ||
		!checksumMatches(request.AdditionalAuthenticatedData,
			request.AdditionalAuthenticatedDataCrc32C) ||
		!checksumMatches(request.InitializationVector, request.InitializationVectorCrc32C) {
		t.Fatal("raw decrypt request did not preserve the envelope contract")
	}
	if !bytes.Equal(client.response.Plaintext, make([]byte, totpPlaintextBytes)) {
		t.Fatal("KMS response plaintext was not wiped")
	}
}

func TestDecryptTOTPRejectsCrossKeyAndUnverifiedResponse(t *testing.T) {
	client := &decryptClientFixture{response: validResponse(bytes.Repeat([]byte{1}, 32))}
	decryptor := newFixtureDecryptor(t, client)
	other := "gcp-kms://projects/harpy-prod/locations/europe-west1/" +
		"keyRings/iam/cryptoKeys/other/cryptoKeyVersions/1"
	_, err := decryptor.DecryptTOTP(
		context.Background(), other, make([]byte, 12), make([]byte, 32),
		make([]byte, 16), [32]byte{1},
	)
	if !errors.Is(err, ErrContract) || client.request != nil {
		t.Fatal("cross-key decrypt was not rejected before the provider call")
	}
	client.response.VerifiedCiphertextCrc32C = false
	_, err = decryptor.DecryptTOTP(
		context.Background(), retiredFixtureKey, make([]byte, 12), make([]byte, 32),
		make([]byte, 16), [32]byte{1},
	)
	if !errors.Is(err, ErrIntegrity) {
		t.Fatalf("unverified response accepted: %v", err)
	}
}

func TestDecryptorCloseIsIdempotentAndFailsClosed(t *testing.T) {
	client := &decryptClientFixture{}
	decryptor := newFixtureDecryptor(t, client)
	if err := decryptor.Close(); err != nil {
		t.Fatalf("close decryptor: %v", err)
	}
	if err := decryptor.Close(); err != nil || client.closed != 1 {
		t.Fatalf("close was not idempotent: calls=%d err=%v", client.closed, err)
	}
	_, err := decryptor.DecryptTOTP(
		context.Background(), activeFixtureKey, make([]byte, 12), make([]byte, 32),
		make([]byte, 16), [32]byte{1},
	)
	if !errors.Is(err, ErrContract) {
		t.Fatalf("closed decryptor accepted work: %v", err)
	}
}

func newFixtureDecryptor(t *testing.T, client rawDecryptClient) *TOTPDecryptor {
	t.Helper()
	decryptor, err := openTOTPDecryptor(
		context.Background(), activeFixtureKey,
		func(context.Context) (rawDecryptClient, error) { return client, nil },
	)
	if err != nil {
		t.Fatalf("open decryptor: %v", err)
	}
	return decryptor
}

func validResponse(plaintext []byte) *kmspb.RawDecryptResponse {
	return &kmspb.RawDecryptResponse{
		Plaintext: bytes.Clone(plaintext), PlaintextCrc32C: checksum(plaintext),
		VerifiedCiphertextCrc32C:                  true,
		VerifiedAdditionalAuthenticatedDataCrc32C: true,
		VerifiedInitializationVectorCrc32C:        true,
	}
}

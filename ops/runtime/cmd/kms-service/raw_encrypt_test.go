package main

import (
	"bytes"
	"context"
	"testing"

	"cloud.google.com/go/kms/apiv1/kmspb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestRawEncryptReturnsGoogleCompatibleIntegrityFields(t *testing.T) {
	request := validEncryptFixture()
	response, err := newKMSService(fixtureKey).RawEncrypt(context.Background(), request)
	if err != nil {
		t.Fatalf("RawEncrypt: %v", err)
	}
	if response.Name != fixtureResource || response.TagLength != gcmTagBytes ||
		len(response.Ciphertext) != len(fixturePlaintext)+gcmTagBytes ||
		!bytes.Equal(response.InitializationVector, fixtureIV) {
		t.Fatalf("unexpected encryption response: %+v", response)
	}
	if !response.VerifiedPlaintextCrc32C ||
		!response.VerifiedAdditionalAuthenticatedDataCrc32C ||
		!response.VerifiedInitializationVectorCrc32C ||
		!validChecksum(response.Ciphertext, response.CiphertextCrc32C) ||
		!validChecksum(response.InitializationVector, response.InitializationVectorCrc32C) {
		t.Fatal("response integrity fields are incomplete")
	}
}

func TestRawEncryptRejectsMissingOrMismatchedIntegrity(t *testing.T) {
	tests := map[string]func(){
		"plaintext": func() {
			request := validEncryptFixture()
			request.PlaintextCrc32C = nil
			assertEncryptInvalid(t, request)
		},
		"aad": func() {
			request := validEncryptFixture()
			request.AdditionalAuthenticatedDataCrc32C.Value++
			assertEncryptInvalid(t, request)
		},
		"iv": func() {
			request := validEncryptFixture()
			request.InitializationVectorCrc32C.Value++
			assertEncryptInvalid(t, request)
		},
	}
	for name, execute := range tests {
		t.Run(name, func(t *testing.T) { execute() })
	}
}

func assertEncryptInvalid(t *testing.T, request *kmspb.RawEncryptRequest) {
	t.Helper()
	_, err := newKMSService(fixtureKey).RawEncrypt(context.Background(), request)
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("RawEncrypt status = %v, want InvalidArgument", status.Code(err))
	}
}

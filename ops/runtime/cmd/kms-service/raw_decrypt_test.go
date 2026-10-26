package main

import (
	"bytes"
	"context"
	"testing"

	"cloud.google.com/go/kms/apiv1/kmspb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestRawDecryptAuthenticatesAndReturnsVerifiedPlaintext(t *testing.T) {
	response, err := newKMSService(fixtureKey).RawDecrypt(
		context.Background(), validDecryptFixture(t),
	)
	if err != nil {
		t.Fatalf("RawDecrypt: %v", err)
	}
	if !bytes.Equal(response.Plaintext, fixturePlaintext) ||
		!validChecksum(response.Plaintext, response.PlaintextCrc32C) ||
		!response.VerifiedCiphertextCrc32C ||
		!response.VerifiedAdditionalAuthenticatedDataCrc32C ||
		!response.VerifiedInitializationVectorCrc32C {
		t.Fatalf("unexpected decryption response: %+v", response)
	}
}

func TestRawDecryptRejectsBadTagLengthAndAuthentication(t *testing.T) {
	wrongLength := validDecryptFixture(t)
	wrongLength.TagLength = 12
	assertDecryptCode(t, wrongLength, codes.InvalidArgument)

	tests := map[string]func(*kmspb.RawDecryptRequest){
		"ciphertext": func(request *kmspb.RawDecryptRequest) {
			request.Ciphertext[0] ^= 0xff
			request.CiphertextCrc32C = checksum(request.Ciphertext)
		},
		"aad": func(request *kmspb.RawDecryptRequest) {
			request.AdditionalAuthenticatedData[0] ^= 0xff
			request.AdditionalAuthenticatedDataCrc32C = checksum(
				request.AdditionalAuthenticatedData,
			)
		},
		"iv": func(request *kmspb.RawDecryptRequest) {
			request.InitializationVector[0] ^= 0xff
			request.InitializationVectorCrc32C = checksum(request.InitializationVector)
		},
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			request := validDecryptFixture(t)
			mutate(request)
			assertDecryptCode(t, request, codes.InvalidArgument)
		})
	}
}

func TestRawDecryptRejectsMismatchedIntegrity(t *testing.T) {
	tests := map[string]func(*kmspb.RawDecryptRequest){
		"ciphertext": func(request *kmspb.RawDecryptRequest) {
			request.CiphertextCrc32C.Value++
		},
		"aad": func(request *kmspb.RawDecryptRequest) {
			request.AdditionalAuthenticatedDataCrc32C = nil
		},
		"iv": func(request *kmspb.RawDecryptRequest) {
			request.InitializationVectorCrc32C.Value++
		},
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			request := validDecryptFixture(t)
			mutate(request)
			assertDecryptCode(t, request, codes.InvalidArgument)
		})
	}
}

func assertDecryptCode(t *testing.T, request *kmspb.RawDecryptRequest, expected codes.Code) {
	t.Helper()
	_, err := newKMSService(fixtureKey).RawDecrypt(context.Background(), request)
	if status.Code(err) != expected {
		t.Fatalf("RawDecrypt status = %v, want %v", status.Code(err), expected)
	}
}

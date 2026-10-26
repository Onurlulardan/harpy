package main

import (
	"context"
	"crypto/aes"
	"crypto/cipher"

	"cloud.google.com/go/kms/apiv1/kmspb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (service *kmsService) RawDecrypt(
	ctx context.Context,
	request *kmspb.RawDecryptRequest,
) (*kmspb.RawDecryptResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, status.FromContextError(err).Err()
	}
	if !service.validDecryptRequest(request) {
		return nil, status.Error(codes.InvalidArgument, "invalid raw decrypt request")
	}
	key := service.keyFor(request.Name)
	defer clear(key[:])
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, status.Error(codes.Internal, "cipher initialization failed")
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, status.Error(codes.Internal, "cipher initialization failed")
	}
	plaintext, err := aead.Open(
		nil,
		request.InitializationVector,
		request.Ciphertext,
		request.AdditionalAuthenticatedData,
	)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "ciphertext authentication failed")
	}
	return &kmspb.RawDecryptResponse{
		Plaintext:                plaintext,
		PlaintextCrc32C:          checksum(plaintext),
		VerifiedCiphertextCrc32C: true,
		VerifiedAdditionalAuthenticatedDataCrc32C: true,
		VerifiedInitializationVectorCrc32C:        true,
		ProtectionLevel:                           kmspb.ProtectionLevel_SOFTWARE,
	}, nil
}

func (service *kmsService) validDecryptRequest(request *kmspb.RawDecryptRequest) bool {
	return request != nil && service.validResource(request.Name) &&
		len(request.Ciphertext) >= gcmTagBytes &&
		len(request.Ciphertext) <= maxPayloadSize+gcmTagBytes &&
		len(request.AdditionalAuthenticatedData) <= maxPayloadSize &&
		len(request.InitializationVector) == gcmNonceBytes &&
		(request.TagLength == 0 || request.TagLength == gcmTagBytes) &&
		validChecksum(request.Ciphertext, request.CiphertextCrc32C) &&
		validChecksum(
			request.AdditionalAuthenticatedData,
			request.AdditionalAuthenticatedDataCrc32C,
		) && validChecksum(
		request.InitializationVector,
		request.InitializationVectorCrc32C,
	)
}

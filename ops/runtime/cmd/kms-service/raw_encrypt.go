package main

import (
	"context"
	"crypto/aes"
	"crypto/cipher"

	"cloud.google.com/go/kms/apiv1/kmspb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (service *kmsService) RawEncrypt(
	ctx context.Context,
	request *kmspb.RawEncryptRequest,
) (*kmspb.RawEncryptResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, status.FromContextError(err).Err()
	}
	if !service.validEncryptRequest(request) {
		return nil, status.Error(codes.InvalidArgument, "invalid raw encrypt request")
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
	ciphertext := aead.Seal(
		nil,
		request.InitializationVector,
		request.Plaintext,
		request.AdditionalAuthenticatedData,
	)
	return &kmspb.RawEncryptResponse{
		Name:                       request.Name,
		Ciphertext:                 ciphertext,
		InitializationVector:       append([]byte(nil), request.InitializationVector...),
		TagLength:                  gcmTagBytes,
		CiphertextCrc32C:           checksum(ciphertext),
		InitializationVectorCrc32C: checksum(request.InitializationVector),
		VerifiedPlaintextCrc32C:    true,
		VerifiedAdditionalAuthenticatedDataCrc32C: true,
		VerifiedInitializationVectorCrc32C:        true,
		ProtectionLevel:                           kmspb.ProtectionLevel_SOFTWARE,
	}, nil
}

func (service *kmsService) validEncryptRequest(request *kmspb.RawEncryptRequest) bool {
	return request != nil && service.validResource(request.Name) &&
		len(request.Plaintext) > 0 && len(request.Plaintext) <= maxPayloadSize &&
		len(request.AdditionalAuthenticatedData) <= maxPayloadSize &&
		len(request.InitializationVector) == gcmNonceBytes &&
		validChecksum(request.Plaintext, request.PlaintextCrc32C) &&
		validChecksum(
			request.AdditionalAuthenticatedData,
			request.AdditionalAuthenticatedDataCrc32C,
		) && validChecksum(
		request.InitializationVector,
		request.InitializationVectorCrc32C,
	)
}

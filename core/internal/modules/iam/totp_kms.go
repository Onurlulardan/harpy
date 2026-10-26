package iam

import (
	"context"

	cryptography "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/cryptography"
	gcpkms "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/gcpkms"
)

type TOTPDecryptor interface {
	cryptography.TOTPSecretDecryptor
	Close() error
}

type TOTPSealer = gcpkms.TOTPSealer
type TOTPEnvelope = gcpkms.TOTPEnvelope

func OpenTOTPDecryptor(
	ctx context.Context,
	activeKeyID string,
) (TOTPDecryptor, error) {
	return gcpkms.OpenTOTPDecryptor(ctx, activeKeyID)
}

func OpenTOTPSealer(ctx context.Context, activeKeyID string) (*TOTPSealer, error) {
	return gcpkms.OpenTOTPSealer(ctx, activeKeyID)
}

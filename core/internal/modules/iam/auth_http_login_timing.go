package iam

import (
	"fmt"

	cryptography "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/cryptography"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authentication"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/password"
)

func newHTTPLoginEqualizer() (
	cryptography.Password, authentication.Argon2idCredential, error,
) {
	provider := cryptography.NewPassword()
	secret, err := password.NewVerificationSecret(
		[]byte("harpy-auth-timing-equalizer-v1"),
	)
	if err != nil {
		return cryptography.Password{}, nil, fmt.Errorf(
			"build IAM login timing secret: %w", err,
		)
	}
	defer secret.Destroy()
	credential, err := provider.Hash(secret)
	if err != nil {
		return cryptography.Password{}, nil, fmt.Errorf(
			"build IAM login timing credential: %w", err,
		)
	}
	return provider, credential, nil
}

func (application *authHTTPApplication) equalizeLogin(candidate []byte) error {
	secret, err := password.NewVerificationSecret(candidate)
	if err != nil {
		return ErrAuthHTTPInvalidCredentials
	}
	defer secret.Destroy()
	_, err = application.password.VerifyArgon2id(secret, application.dummyPassword)
	if err != nil {
		return ErrAuthHTTPUnavailable
	}
	return nil
}

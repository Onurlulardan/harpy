package cryptography

import (
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authentication"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/password"
)

type Password struct{ hasher password.Hasher }

func NewPassword() Password { return Password{hasher: password.NewHasher()} }

func (provider Password) Hash(
	secret authentication.PasswordSecret,
) (authentication.Argon2idCredential, error) {
	concrete, ok := secret.(*password.Secret)
	if !ok || concrete == nil {
		return nil, authentication.ErrAuthenticationContract
	}
	credential, err := provider.hasher.Hash(concrete)
	if err != nil {
		return nil, err
	}
	return credential, nil
}

func (provider Password) VerifyArgon2id(
	secret authentication.PasswordSecret,
	credential authentication.Argon2idCredential,
) (bool, error) {
	concreteSecret, secretOK := secret.(*password.Secret)
	concreteCredential, credentialOK := credential.(password.Argon2idCredential)
	if !secretOK || concreteSecret == nil || !credentialOK {
		return false, authentication.ErrAuthenticationContract
	}
	return password.Verify(concreteSecret, concreteCredential)
}

func (provider Password) VerifyLegacy(
	secret authentication.PasswordSecret,
	credential authentication.LegacyBcryptCredential,
	now time.Time,
) (bool, error) {
	concreteSecret, secretOK := secret.(*password.Secret)
	concreteCredential, credentialOK := credential.(password.LegacyBcryptCredential)
	if !secretOK || concreteSecret == nil || !credentialOK {
		return false, authentication.ErrAuthenticationContract
	}
	return password.VerifyLegacy(concreteSecret, concreteCredential, now)
}

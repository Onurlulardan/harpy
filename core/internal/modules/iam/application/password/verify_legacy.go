package password

import (
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func VerifyLegacy(
	secret *Secret,
	credential LegacyBcryptCredential,
	now time.Time,
) (bool, error) {
	if !secret.usable() {
		return false, ErrSecretDestroyed
	}
	if !credential.Valid() || now.IsZero() {
		return false, ErrCredentialInvalid
	}
	err := bcrypt.CompareHashAndPassword(credential.hash[:], secret.value)
	switch {
	case err == nil && !now.Before(credential.RetirementDeadline()):
		return false, ErrLegacyCredentialExpired
	case err == nil:
		return true, nil
	case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
		return false, nil
	default:
		return false, ErrCredentialInvalid
	}
}

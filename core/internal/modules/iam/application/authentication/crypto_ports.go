package authentication

import (
	"time"

	"github.com/google/uuid"
)

type PasswordCryptography interface {
	Hash(PasswordSecret) (Argon2idCredential, error)
	VerifyArgon2id(PasswordSecret, Argon2idCredential) (bool, error)
	VerifyLegacy(PasswordSecret, LegacyBcryptCredential, time.Time) (bool, error)
}

type RefreshGenerator interface {
	Generate() (IssuedRefresh, error)
}

type RefreshKeyring interface {
	Active() (string, [32]byte, error)
	HMACKey(string) ([32]byte, error)
}

type AccessIssuer interface {
	Issue(AccessClaims) (string, error)
}

type UUIDSource func() (uuid.UUID, error)
type Clock func() time.Time

func SystemUUID() (uuid.UUID, error) { return uuid.NewRandom() }

func SystemClock() time.Time { return time.Now().UTC() }

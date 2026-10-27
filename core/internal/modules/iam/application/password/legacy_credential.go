package password

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

const (
	AlgorithmBcryptV1Adopted = "bcrypt_v1_adopted"
	LegacyBcryptBytes        = 60
	LegacyRetirementDays     = 180
	LegacyMaximumCost        = 16
)

type LegacyBcryptCredential struct {
	hash      [LegacyBcryptBytes]byte
	adoptedAt time.Time
	issued    bool
}

func LegacyBcryptCredentialFromStorage(
	hash string,
	adoptedAt time.Time,
) (LegacyBcryptCredential, error) {
	if len(hash) != LegacyBcryptBytes || adoptedAt.IsZero() {
		return LegacyBcryptCredential{}, ErrCredentialInvalid
	}
	cost, err := bcrypt.Cost([]byte(hash))
	if err != nil || cost < bcrypt.MinCost || cost > LegacyMaximumCost {
		return LegacyBcryptCredential{}, ErrCredentialInvalid
	}
	credential := LegacyBcryptCredential{adoptedAt: adoptedAt.UTC(), issued: true}
	copy(credential.hash[:], hash)
	return credential, nil
}

func (credential LegacyBcryptCredential) Algorithm() string {
	return AlgorithmBcryptV1Adopted
}

func (credential LegacyBcryptCredential) AdoptedAt() time.Time {
	return credential.adoptedAt
}

func (credential LegacyBcryptCredential) RetirementDeadline() time.Time {
	if !credential.Valid() {
		return time.Time{}
	}
	return credential.adoptedAt.AddDate(0, 0, LegacyRetirementDays)
}

func (credential LegacyBcryptCredential) Valid() bool {
	return credential.issued && !credential.adoptedAt.IsZero() &&
		credential.hash != [LegacyBcryptBytes]byte{}
}

func (credential LegacyBcryptCredential) NeedsRehash() bool { return credential.Valid() }

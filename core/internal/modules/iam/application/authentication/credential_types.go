package authentication

import "time"

const (
	AlgorithmArgon2idV1      = "argon2id_v1"
	AlgorithmBcryptV1Adopted = "bcrypt_v1_adopted"
)

type PasswordSecret interface {
	Destroy()
}

type Argon2idCredential interface {
	Valid() bool
	Salt() [16]byte
	Tag() [32]byte
	MemoryKiB() uint32
	Iterations() uint32
	Parallelism() uint8
}

type LegacyBcryptCredential interface {
	Valid() bool
	AdoptedAt() time.Time
	RetirementDeadline() time.Time
}

type PasswordCredential struct {
	Algorithm string
	Argon2id  Argon2idCredential
	Legacy    LegacyBcryptCredential
}

func (credential PasswordCredential) invalid() bool {
	switch credential.Algorithm {
	case AlgorithmArgon2idV1:
		return nilInterface(credential.Argon2id) || !credential.Argon2id.Valid() ||
			!nilInterface(credential.Legacy)
	case AlgorithmBcryptV1Adopted:
		return !nilInterface(credential.Argon2id) || nilInterface(credential.Legacy) ||
			!credential.Legacy.Valid()
	default:
		return true
	}
}

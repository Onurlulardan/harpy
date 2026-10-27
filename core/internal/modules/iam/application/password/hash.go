package password

import (
	"crypto/rand"
	"io"

	"golang.org/x/crypto/argon2"
)

type Hasher struct{ randomness io.Reader }

func NewHasher() Hasher { return Hasher{randomness: rand.Reader} }

func (hasher Hasher) Hash(secret *Secret) (Argon2idCredential, error) {
	if !secret.usable() || hasher.randomness == nil {
		return Argon2idCredential{}, ErrSecretDestroyed
	}
	var salt [SaltBytesV1]byte
	if _, err := io.ReadFull(hasher.randomness, salt[:]); err != nil {
		return Argon2idCredential{}, err
	}
	tag := argon2.IDKey(
		secret.value, salt[:], IterationsV1, MemoryKiBV1, ParallelismV1, TagBytesV1,
	)
	credential, err := Argon2idCredentialFromStorage(
		salt[:], tag, MemoryKiBV1, IterationsV1, ParallelismV1,
	)
	wipe(tag)
	return credential, err
}

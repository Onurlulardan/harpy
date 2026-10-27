package password

import (
	"crypto/subtle"

	"golang.org/x/crypto/argon2"
)

func Verify(secret *Secret, credential Argon2idCredential) (bool, error) {
	if !secret.usable() {
		return false, ErrSecretDestroyed
	}
	if !credential.Valid() {
		return false, ErrCredentialInvalid
	}
	actual := argon2.IDKey(
		secret.value, credential.salt[:], credential.iterations,
		credential.memoryKiB, credential.parallelism, TagBytesV1,
	)
	matched := subtle.ConstantTimeCompare(actual, credential.tag[:]) == 1
	wipe(actual)
	return matched, nil
}

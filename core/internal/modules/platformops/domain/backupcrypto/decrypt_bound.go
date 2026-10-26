package backupcrypto

import (
	"bufio"
	"io"
)

func DecryptBound(
	ciphertext io.Reader,
	plaintext io.Writer,
	expected BoundContext,
	resolver OpenKeyResolver,
) (int64, Context, error) {
	if ciphertext == nil || plaintext == nil || resolver == nil ||
		!validBoundContext(expected) {
		return 0, Context{}, ErrInvalidArchive
	}
	reader := bufio.NewReader(ciphertext)
	headerBytes, header, err := readPreamble(reader)
	if err != nil {
		return 0, Context{}, err
	}
	actual := header.context()
	if actual.EnvironmentID != expected.EnvironmentID ||
		actual.RecoverySetID != expected.RecoverySetID ||
		actual.ComponentID != expected.ComponentID {
		return 0, Context{}, ErrInvalidArchive
	}
	keys, found := resolver(actual.DRKeyID, actual.RecoveryKeyID)
	if !found {
		return 0, Context{}, ErrInvalidArchive
	}
	defer clear(keys.DRKEK)
	defer clear(keys.RecoveryPrivate)
	count, err := decryptBody(reader, plaintext, headerBytes, header, keys)
	if err != nil {
		return 0, Context{}, err
	}
	return count, actual, nil
}

func validBoundContext(value BoundContext) bool {
	return validUUID(value.EnvironmentID) && validUUID(value.RecoverySetID) &&
		componentPattern.MatchString(value.ComponentID)
}

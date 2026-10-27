package fcartv1

import (
	"errors"
	"testing"
)

func TestDecryptRejectsChunkReorderDuplicateAndNonceReuse(t *testing.T) {
	plaintext := make([]byte, ChunkPlaintextBytes+7)
	for index := range plaintext {
		plaintext[index] = byte(index)
	}
	ciphertext, result := encryptFixture(t, plaintext)
	firstStart := HeaderBytes
	secondStart := firstStart + ChunkFramingBytes + ChunkPlaintextBytes
	firstFrame := cloneBytes(ciphertext[firstStart:secondStart])
	secondFrame := cloneBytes(ciphertext[secondStart:])

	reordered := cloneBytes(ciphertext[:HeaderBytes])
	reordered = append(reordered, secondFrame...)
	reordered = append(reordered, firstFrame...)

	duplicated := cloneBytes(ciphertext[:HeaderBytes])
	duplicated = append(duplicated, firstFrame...)
	duplicated = append(duplicated, firstFrame...)

	reusedNonce := cloneBytes(ciphertext)
	copy(
		reusedNonce[secondStart+12:secondStart+12+NonceBytes],
		reusedNonce[firstStart+12:firstStart+12+NonceBytes],
	)

	for _, test := range []struct {
		name       string
		ciphertext []byte
		wanted     error
	}{
		{"reordered", reordered, ErrSequence},
		{"duplicated", duplicated, ErrSequence},
		{"nonce-reused", reusedNonce, ErrNonceReuse},
	} {
		t.Run(test.name, func(t *testing.T) {
			metadata := metadataForBytes(result.Header, test.ciphertext)
			_, _, err := decryptFixture(t, test.ciphertext, metadata)
			if !errors.Is(err, test.wanted) {
				t.Fatalf("error = %v", err)
			}
		})
	}
}

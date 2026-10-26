package fcartv1

import (
	"errors"
	"testing"
)

func TestDecryptRejectsHeaderSubstitution(t *testing.T) {
	ciphertext, result := encryptFixture(t, []byte("secret"))
	for name, offset := range map[string]int{
		"artifact": len(Version),
		"attempt":  len(Version) + 16 + 7,
		"format":   len(Version) + 16 + 8,
	} {
		t.Run(name, func(t *testing.T) {
			changed := cloneBytes(ciphertext)
			if name == "format" {
				changed[offset] = 2
			} else {
				changed[offset] ^= 0x80
			}
			metadata := metadataForBytes(result.Header, changed)
			_, _, err := decryptFixture(t, changed, metadata)
			if !errors.Is(err, ErrIdentityMismatch) {
				t.Fatalf("header substitution error = %v", err)
			}
		})
	}
}

func TestDecryptRejectsAuthenticatedChunkTamper(t *testing.T) {
	ciphertext, result := encryptFixture(t, []byte("secret"))
	changed := cloneBytes(ciphertext)
	changed[len(changed)-1] ^= 1
	metadata := metadataForBytes(result.Header, changed)
	_, _, err := decryptFixture(t, changed, metadata)
	if !errors.Is(err, ErrAuthentication) {
		t.Fatalf("tamper error = %v", err)
	}
}

func TestDecryptRejectsWrongDEK(t *testing.T) {
	ciphertext, result := encryptFixture(t, []byte("secret"))
	wrong := fixtureDEK(t)
	wrong[0] ^= 1
	var destination shortBuffer
	_, err := Decrypt(
		&destination, newByteReader(ciphertext), result.CiphertextMetadata, wrong,
	)
	if !errors.Is(err, ErrAuthentication) {
		t.Fatalf("wrong-key error = %v", err)
	}
}

func TestDecryptRejectsChecksumMismatch(t *testing.T) {
	ciphertext, result := encryptFixture(t, []byte("secret"))
	metadata := result.CiphertextMetadata
	metadata.CiphertextSHA256[0] ^= 1
	_, _, err := decryptFixture(t, ciphertext, metadata)
	if !errors.Is(err, ErrChecksum) {
		t.Fatalf("checksum error = %v", err)
	}
}

func TestDecryptFailsClosedOnPlaintextShortWrite(t *testing.T) {
	ciphertext, result := encryptFixture(t, []byte("secret"))
	destination := &shortBuffer{short: true}
	_, err := Decrypt(
		destination, newByteReader(ciphertext), result.CiphertextMetadata,
		fixtureDEK(t),
	)
	if !errors.Is(err, ErrShortWrite) {
		t.Fatalf("plaintext short-write error = %v", err)
	}
}

func TestDecryptRejectsTruncationAndTrailingBytes(t *testing.T) {
	ciphertext, result := encryptFixture(t, []byte("secret"))
	for _, test := range []struct {
		name    string
		changed []byte
		wanted  error
	}{
		{"truncated", cloneBytes(ciphertext[:len(ciphertext)-1]), ErrTruncated},
		{"trailing", append(cloneBytes(ciphertext), 0), ErrTrailingData},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, _, err := decryptFixture(t, test.changed, result.CiphertextMetadata)
			if !errors.Is(err, test.wanted) {
				t.Fatalf("error = %v", err)
			}
		})
	}
}

package fcartv1

import (
	"bytes"
	"errors"
	"testing"
)

func TestEncryptRejectsNonceReuse(t *testing.T) {
	plaintext := make([]byte, ChunkPlaintextBytes+1)
	var nonce [NonceBytes]byte
	nonce[0] = 1
	var destination bytes.Buffer
	_, err := Encrypt(
		&destination, bytes.NewReader(plaintext), fixtureHeader(t), fixtureDEK(t),
		&deterministicNonceSource{fixed: &nonce},
	)
	if !errors.Is(err, ErrNonceReuse) {
		t.Fatalf("nonce reuse error = %v", err)
	}
	wantBytes := HeaderBytes + ChunkFramingBytes + ChunkPlaintextBytes
	if destination.Len() != wantBytes {
		t.Fatalf("second frame was partially written: %d", destination.Len())
	}
}

func TestEncryptRejectsRandomSourceFailure(t *testing.T) {
	randomFailure := errors.New("entropy unavailable")
	var destination bytes.Buffer
	_, err := Encrypt(
		&destination, bytes.NewReader([]byte("secret")), fixtureHeader(t),
		fixtureDEK(t), &deterministicNonceSource{err: randomFailure},
	)
	if !errors.Is(err, ErrRandomSource) || destination.Len() != HeaderBytes {
		t.Fatalf("random-source failure = %d bytes, %v", destination.Len(), err)
	}
}

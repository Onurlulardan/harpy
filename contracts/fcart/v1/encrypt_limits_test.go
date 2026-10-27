package fcartv1

import (
	"bytes"
	"errors"
	"io"
	"testing"
)

func TestEncryptEnforcesCiphertextLimitBeforeFrameWrite(t *testing.T) {
	plaintext := []byte("bounded")
	limit := uint64(HeaderBytes + ChunkFramingBytes + len(plaintext) - 1)
	var destination bytes.Buffer
	random := &deterministicNonceSource{seed: 1}
	_, err := encryptWithLimit(
		&destination, bytes.NewReader(plaintext), fixtureHeader(t), fixtureDEK(t),
		random, limit,
	)
	if !errors.Is(err, ErrCiphertextLimit) {
		t.Fatalf("limit error = %v", err)
	}
	if destination.Len() != HeaderBytes {
		t.Fatalf("overflowing frame emitted %d bytes", destination.Len())
	}
	if random.sequence != 0 {
		t.Fatal("nonce consumed before frame-size preflight")
	}
}

func TestEncryptAcceptsExactCiphertextBoundary(t *testing.T) {
	plaintext := []byte("bounded")
	limit := uint64(HeaderBytes + ChunkFramingBytes + len(plaintext))
	var destination bytes.Buffer
	result, err := encryptWithLimit(
		&destination, bytes.NewReader(plaintext), fixtureHeader(t), fixtureDEK(t),
		&deterministicNonceSource{}, limit,
	)
	if err != nil || result.CiphertextBytes != limit {
		t.Fatalf("exact boundary = %#v, %v", result, err)
	}
}

func TestEncryptRejectsLimitBeforeHeaderWrite(t *testing.T) {
	var destination bytes.Buffer
	_, err := encryptWithLimit(
		&destination, bytes.NewReader(nil), fixtureHeader(t), fixtureDEK(t),
		&deterministicNonceSource{}, uint64(HeaderBytes-1),
	)
	if !errors.Is(err, ErrCiphertextLimit) || destination.Len() != 0 {
		t.Fatalf("header limit = %d bytes, %v", destination.Len(), err)
	}
}

func TestEncryptFailsClosedOnShortDestinationWrite(t *testing.T) {
	destination := &shortBuffer{short: true}
	_, err := Encrypt(
		destination, bytes.NewReader([]byte("secret")), fixtureHeader(t),
		fixtureDEK(t), &deterministicNonceSource{},
	)
	if !errors.Is(err, ErrShortWrite) {
		t.Fatalf("short-write error = %v", err)
	}
}

func TestEncryptFailsClosedOnChunkShortWrite(t *testing.T) {
	destination := &shortAfterWriter{fullWrites: 1}
	_, err := Encrypt(
		destination, bytes.NewReader([]byte("secret")), fixtureHeader(t),
		fixtureDEK(t), &deterministicNonceSource{},
	)
	if !errors.Is(err, ErrShortWrite) {
		t.Fatalf("chunk short-write error = %v", err)
	}
}

func TestEncryptFailsClosedOnPlaintextReadError(t *testing.T) {
	readFailure := errors.New("source failed")
	reader := io.MultiReader(bytes.NewReader([]byte("partial")), errorReader{readFailure})
	var destination bytes.Buffer
	_, err := Encrypt(
		&destination, reader, fixtureHeader(t), fixtureDEK(t),
		&deterministicNonceSource{},
	)
	if !errors.Is(err, ErrPlaintextRead) || destination.Len() != HeaderBytes {
		t.Fatalf("plaintext failure = %d bytes, %v", destination.Len(), err)
	}
}

type errorReader struct{ err error }

func (reader errorReader) Read([]byte) (int, error) { return 0, reader.err }

type shortAfterWriter struct {
	bytes.Buffer
	fullWrites int
	writes     int
}

func (writer *shortAfterWriter) Write(value []byte) (int, error) {
	writer.writes++
	if writer.writes > writer.fullWrites && len(value) > 0 {
		return writer.Buffer.Write(value[:len(value)-1])
	}
	return writer.Buffer.Write(value)
}

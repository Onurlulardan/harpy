package password

import (
	"bytes"
	"errors"
	"testing"
)

type failingReader struct{ err error }

func (reader failingReader) Read([]byte) (int, error) { return 0, reader.err }

func TestHasherProducesExactArgon2idV1Credential(t *testing.T) {
	secret, err := NewVerificationSecret([]byte("correct horse battery staple"))
	if err != nil {
		t.Fatal(err)
	}
	defer secret.Destroy()
	hasher := Hasher{randomness: bytes.NewReader(bytes.Repeat([]byte{7}, SaltBytesV1))}
	credential, err := hasher.Hash(secret)
	if err != nil {
		t.Fatal(err)
	}
	if !credential.Valid() || credential.NeedsRehash() ||
		credential.Algorithm() != AlgorithmArgon2idV1 ||
		credential.MemoryKiB() != MemoryKiBV1 ||
		credential.Iterations() != IterationsV1 ||
		credential.Parallelism() != ParallelismV1 ||
		credential.Salt() != [SaltBytesV1]byte{7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7} {
		t.Fatalf("unexpected Argon2id profile: %+v", credential)
	}
	matched, err := Verify(secret, credential)
	if err != nil || !matched {
		t.Fatalf("correct password result: matched=%v err=%v", matched, err)
	}
	wrong, err := NewVerificationSecret([]byte("wrong horse battery staple"))
	if err != nil {
		t.Fatal(err)
	}
	defer wrong.Destroy()
	matched, err = Verify(wrong, credential)
	if err != nil || matched {
		t.Fatalf("wrong password result: matched=%v err=%v", matched, err)
	}
}

func TestHasherAndVerifierFailClosed(t *testing.T) {
	secret, err := NewVerificationSecret([]byte("long-enough-password"))
	if err != nil {
		t.Fatal(err)
	}
	readErr := errors.New("random source unavailable")
	if _, err := (Hasher{randomness: failingReader{err: readErr}}).Hash(secret); !errors.Is(err, readErr) {
		t.Fatalf("random source error = %v", err)
	}
	secret.Destroy()
	if _, err := NewHasher().Hash(secret); !errors.Is(err, ErrSecretDestroyed) {
		t.Fatalf("destroyed hash error = %v", err)
	}
	if _, err := Verify(secret, Argon2idCredential{}); !errors.Is(err, ErrSecretDestroyed) {
		t.Fatalf("destroyed verify error = %v", err)
	}
}

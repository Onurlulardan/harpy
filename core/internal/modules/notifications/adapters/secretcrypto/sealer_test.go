package secretcrypto

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
)

func testSealer(t *testing.T) *Sealer {
	t.Helper()
	keys := map[int64][32]byte{1: {1}, 2: {2}}
	random := bytes.NewReader(bytes.Repeat([]byte{7}, 512))
	wrapper, err := NewAESKeyWrapper(keys, random)
	if err != nil {
		t.Fatal(err)
	}
	sealer, err := New(random, wrapper)
	if err != nil {
		t.Fatal(err)
	}
	return sealer
}

func TestSecretSealerAuthenticatesAndRewrapsWithoutChangingCiphertext(t *testing.T) {
	id, _ := domain.ParseID("11111111-1111-4111-8111-111111111111")
	plaintext := []byte("smtp-password-that-must-never-be-logged")
	sealer := testSealer(t)
	sealed, err := sealer.Seal(context.Background(), id, plaintext, 1)
	if err != nil || bytes.Contains(sealed.Ciphertext, plaintext) {
		t.Fatalf("Seal() err=%v plaintext leaked=%t", err,
			bytes.Contains(sealed.Ciphertext, plaintext))
	}
	rewrapped, err := sealer.Rewrap(context.Background(), sealed, 2)
	if err != nil || !bytes.Equal(rewrapped.Ciphertext, sealed.Ciphertext) ||
		bytes.Equal(rewrapped.WrappedDEK, sealed.WrappedDEK) {
		t.Fatalf("Rewrap() err=%v", err)
	}
	opened, err := sealer.Open(context.Background(), rewrapped)
	if err != nil || !bytes.Equal(opened, plaintext) {
		t.Fatalf("Open() = %q, %v", opened, err)
	}
}

func TestSecretSealerRejectsTamperWithoutLeakingPlaintext(t *testing.T) {
	id, _ := domain.ParseID("11111111-1111-4111-8111-111111111111")
	plaintext := []byte("unique-sensitive-value")
	sealer := testSealer(t)
	sealed, _ := sealer.Seal(context.Background(), id, plaintext, 1)
	sealed.Ciphertext[0] ^= 1
	_, err := sealer.Open(context.Background(), sealed)
	if err == nil || strings.Contains(err.Error(), string(plaintext)) {
		t.Fatalf("tamper error = %v", err)
	}
}

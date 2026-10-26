package passwordblocklistv1

import (
	"crypto/sha256"
	"errors"
	"testing"
)

func TestEncodeCanonicalizesOrderAndRejectsDuplicates(t *testing.T) {
	a := sha256.Sum256([]byte("a"))
	b := sha256.Sum256([]byte("b"))
	first, err := Encode(KindCompromised, [][32]byte{b, a})
	if err != nil {
		t.Fatal(err)
	}
	second, err := Encode(KindCompromised, [][32]byte{a, b})
	if err != nil {
		t.Fatal(err)
	}
	if string(first) != string(second) {
		t.Fatal("equivalent input did not produce canonical bytes")
	}
	if _, err := Encode(KindCommon, [][32]byte{a, a}); !errors.Is(err, ErrInvalidArtifact) {
		t.Fatalf("duplicate digest error=%v", err)
	}
	if _, err := Encode(0, [][32]byte{a}); !errors.Is(err, ErrInvalidArtifact) {
		t.Fatalf("invalid kind error=%v", err)
	}
}

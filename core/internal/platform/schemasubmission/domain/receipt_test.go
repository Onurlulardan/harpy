package domain

import (
	"testing"

	"github.com/google/uuid"
)

func TestReceiptRejectsIncompleteEvidence(t *testing.T) {
	id, err := NewJobID(uuid.MustParse("018f4f9e-7c21-7a3b-8c4d-123456789abc"))
	if err != nil {
		t.Fatal(err)
	}
	digest := [32]byte{1}
	if receipt, ok := NewReceipt(id, 1, digest, true); !ok || !receipt.Replay {
		t.Fatal("complete replay receipt rejected")
	}
	if _, ok := NewReceipt(id, 0, digest, false); ok {
		t.Fatal("zero generation accepted")
	}
	if _, ok := NewReceipt(id, 1, [32]byte{}, false); ok {
		t.Fatal("zero digest accepted")
	}
}

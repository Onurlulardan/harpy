package application

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestEvidenceRoundTripIsSealedAndRedacted(t *testing.T) {
	actor := uuid.MustParse("019c0000-0000-7000-8000-000000000101")
	tenant := uuid.MustParse("019c0000-0000-7000-8000-000000000102")
	evidence, err := NewEvidence(actor, tenant, 2, 3, 4,
		"changelog.read", "changelog.list")
	if err != nil || evidence.Digest() == [32]byte{} {
		t.Fatalf("evidence = %#v, err = %v", evidence, err)
	}
	ctx, err := WithEvidence(context.Background(), evidence)
	loaded, ok := FromContext(ctx)
	if err != nil || !ok || loaded.Digest() != evidence.Digest() {
		t.Fatal("sealed audit authority evidence was not preserved")
	}
}

func TestEvidenceAcceptsAdoptedIdentityAndRejectsIncompleteVersion(t *testing.T) {
	legacy := uuid.MustParse("85f17fb4-d9f0-4ed8-99ee-7c079ba863fa")
	v7 := uuid.MustParse("019c0000-0000-7000-8000-000000000103")
	if _, err := NewEvidence(legacy, legacy, 1, 1, 1,
		"changelog.read", "changelog.list"); err != nil {
		t.Fatalf("adopted UUID identity was rejected: %v", err)
	}
	if _, err := NewEvidence(v7, v7, 0, 1, 1,
		"changelog.read", "changelog.list"); err == nil {
		t.Fatal("missing token version was accepted")
	}
}

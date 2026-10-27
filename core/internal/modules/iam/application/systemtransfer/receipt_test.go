package systemtransfer

import (
	"crypto/sha256"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestRowReceiptContainsOnlyRedactedEvidence(t *testing.T) {
	source := sha256.Sum256([]byte("source"))
	result := sha256.Sum256([]byte("result"))
	receipt, err := NewRowReceipt(7, "rejected", "", uuid.Nil, 0,
		"REVISION_CONFLICT", source, result)
	if err != nil || !receipt.Valid() {
		t.Fatalf("receipt=%+v err=%v", receipt, err)
	}
	if strings.Contains(strings.ToLower(receipt.ErrorCode), "source") {
		t.Fatal("receipt contains source value")
	}
}

func TestRowIdempotencyKeyIsDeterministic(t *testing.T) {
	transferID := uuid.MustParse("019be000-0000-7000-8000-000000000001")
	digest := sha256.Sum256([]byte("row"))
	first, err := RowIdempotencyKey(transferID, 9, digest)
	second, secondErr := RowIdempotencyKey(transferID, 9, digest)
	if err != nil || secondErr != nil || first != second || len(first) != 72 {
		t.Fatalf("keys=%q/%q errors=%v/%v", first, second, err, secondErr)
	}
}

func TestJobReferenceCarriesOnlyIdentity(t *testing.T) {
	id := uuid.MustParse("019be000-0000-7000-8000-000000000001")
	reference, err := JobReference("users", id)
	resource, parsed, parseErr := ParseJobReference(reference)
	if err != nil || parseErr != nil || resource != "users" || parsed != id {
		t.Fatalf("reference=%q values=%s/%s errors=%v/%v",
			reference, resource, parsed, err, parseErr)
	}
	digest, digestErr := JobReferenceSHA256(reference)
	secondDigest, secondErr := JobReferenceSHA256(reference)
	if digestErr != nil || secondErr != nil || digest != secondDigest || len(digest) != 64 {
		t.Fatalf("payload digests=%q/%q errors=%v/%v",
			digest, secondDigest, digestErr, secondErr)
	}
}

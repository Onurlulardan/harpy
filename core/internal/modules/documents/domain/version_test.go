package domain

import (
	"crypto/sha256"
	"errors"
	"testing"
	"time"
)

func TestReadyVersionBindsObjectToTenantDocumentAndVersion(t *testing.T) {
	tenant := mustID(t, "11111111-1111-4111-8111-111111111111")
	item := mustID(t, "018f0000-0000-7000-8000-000000000201")
	versionID := mustID(t, "018f0000-0000-7000-8000-000000000202")
	key, _ := ContentObjectKey(tenant, item, versionID)
	declaration := FileDeclaration{Filename: "evidence.pdf", DeclaredMIME: "application/pdf",
		DetectedKind: FilePDF, Size: 7}
	receipt := ObjectReceipt{Key: key, Version: "opaque-version", KeyID: "documents-kek-1",
		Generation: 1, PlaintextBytes: 7, CiphertextBytes: 128,
		PlaintextSHA256:  sha256.Sum256([]byte("content")),
		CiphertextSHA256: sha256.Sum256([]byte("ciphertext")), WrappedDEK: make([]byte, 115)}
	version, err := NewReadyVersion(versionID, tenant, item, 1, declaration,
		"application/pdf", receipt, "clamav-20260718", time.Now().UTC())
	if err != nil || !version.Downloadable() {
		t.Fatalf("version = %+v, %v", version, err)
	}
	receipt.Key = "tenants/foreign/content"
	if _, err = NewReadyVersion(versionID, tenant, item, 1, declaration,
		"application/pdf", receipt, "clamav-20260718", time.Now().UTC()); !errors.Is(err, ErrVersionInvalid) {
		t.Fatalf("foreign key error = %v", err)
	}
}

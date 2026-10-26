package upload

import (
	"crypto/sha256"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

func TestPutReceiptRequiresExactHashKindTenantAndGeneration(t *testing.T) {
	key, _ := domain.ContentObjectKey(uploadTenant, uploadDocument, uploadVersion)
	plan := AttemptPlan{UploadID: uploadHandle, TenantID: uploadTenant, ActorID: uploadActor,
		DocumentID: uploadDocument, CandidateVersionID: uploadVersion,
		BillingReservationID: uploadReservation, EffectReceiptID: uploadReceipt,
		Generation: 1, ObjectKey: key, OriginalName: "evidence.pdf",
		DeclaredMIME: "application/pdf", ExpectedContentSHA256: sha256.Sum256([]byte("content")),
		ExpectedBytes: 7, DeclaredKind: domain.FilePDF}
	receipt := EffectReceipt{Plan: plan, Object: domain.ObjectReceipt{Key: key,
		Version: "opaque-v1", KeyID: "documents-kek-1", Generation: 1,
		PlaintextBytes: 7, CiphertextBytes: 117,
		PlaintextSHA256:  plan.ExpectedContentSHA256,
		CiphertextSHA256: sha256.Sum256([]byte("ciphertext")), WrappedDEK: make([]byte, 115)},
		DetectedKind: domain.FilePDF, DetectedMIME: "application/pdf",
		ScanRevision: "clamav-20260718", ScannedAt: uploadNow}
	if !validEffectReceipt(plan, receipt) {
		t.Fatal("valid receipt rejected")
	}
	receipt.Object.Generation = 2
	if validEffectReceipt(plan, receipt) {
		t.Fatal("wrong generation accepted")
	}
}

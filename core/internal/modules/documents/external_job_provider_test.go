package documents

import (
	"context"
	"crypto/sha256"
	"testing"
	"time"

	effectsv1 "github.com/Onurlulardan/harpy/contracts/effects/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/api/jobcontract"
	adoptionapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/adoption"
	archiveapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/archive"
	deletionapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/deletion"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	uploadreconcileapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/uploadreconcile"
)

var (
	jobTestTenant   = mustDocumentID("11111111-1111-4111-8111-111111111111")
	jobTestResource = mustDocumentID("018f0000-0000-7000-8000-000000000701")
	jobTestID       = mustDocumentID("018f0000-0000-7000-8000-000000000702")
	jobTestNow      = time.Date(2026, 7, 18, 14, 0, 0, 0, time.UTC)
)

type archiveEffectStub struct{ snapshot archiveapp.Snapshot }

func (stub archiveEffectStub) Execute(context.Context, foundation.ID, foundation.ID,
	foundation.ID,
) (archiveapp.Snapshot, error) {
	return stub.snapshot, nil
}

type adoptionCopyEffectStub struct{ checkpoint adoptionapp.Checkpoint }

func (stub adoptionCopyEffectStub) Execute(context.Context, adoptionapp.Authority,
	foundation.ID,
) (adoptionapp.Checkpoint, error) {
	return stub.checkpoint, nil
}

type deletionEffectStub struct{ result deletionapp.Result }

func (stub deletionEffectStub) Execute(context.Context, foundation.ID, foundation.ID,
	foundation.ID,
) (deletionapp.Result, error) {
	return stub.result, nil
}

type uploadReconcileEffectStub struct{ result uploadreconcileapp.Result }

func (stub uploadReconcileEffectStub) Execute(context.Context, foundation.ID, foundation.ID,
	uint64,
) (uploadreconcileapp.Result, error) {
	return stub.result, nil
}

func TestArchiveProviderReturnsDefinitiveReadyReceipt(t *testing.T) {
	digest := sha256.Sum256([]byte("archive-ciphertext"))
	provider, err := newArchiveProvider(archiveEffectStub{snapshot: archiveapp.Snapshot{
		ID: jobTestResource, TenantID: jobTestTenant, State: archiveapp.StateReady,
		CiphertextSHA256: digest,
	}}, func() time.Time { return jobTestNow })
	if err != nil {
		t.Fatal(err)
	}
	request := documentsEffectRequest(t, "archive")
	result := provider.Execute(context.Background(), request)
	receipt, ok := result.Receipt()
	if !ok || receipt.Code != "DOCUMENT_ARCHIVE_READY" ||
		receipt.Status != effectsv1.ReceiptSucceeded {
		t.Fatalf("archive receipt=%+v", receipt)
	}
}

func TestDeletionProviderReturnsExactAbsenceReceipt(t *testing.T) {
	deleteDigest := sha256.Sum256([]byte("delete"))
	absenceDigest := sha256.Sum256([]byte("absent"))
	provider, err := newDeletionProvider(deletionEffectStub{result: deletionapp.Result{
		Task: deletionapp.Task{ID: jobTestResource, TenantID: jobTestTenant},
		DeleteReceipt: deletionapp.DeleteReceipt{ProviderRequestID: "minio-delete-1",
			DeleteSHA256: deleteDigest, AbsenceSHA256: absenceDigest,
			VerifiedAt: jobTestNow}, ConfirmedAt: jobTestNow,
	}}, func() time.Time { return jobTestNow })
	if err != nil {
		t.Fatal(err)
	}
	result := provider.Execute(context.Background(), documentsEffectRequest(t, "deletion"))
	receipt, ok := result.Receipt()
	if !ok || receipt.Code != "DOCUMENT_BLOB_ABSENT" ||
		receipt.ProviderReference == nil || *receipt.ProviderReference != "minio-delete-1" {
		t.Fatalf("deletion receipt=%+v", receipt)
	}
}

func TestUploadReconcileProviderReturnsDefinitiveReceipt(t *testing.T) {
	provider, err := newUploadReconcileProvider(uploadReconcileEffectStub{
		result: uploadreconcileapp.Result{UploadID: jobTestResource,
			State: uploadreconcileapp.StateRetryable, ObjectFound: true,
			DeletionID: jobTestID},
	}, func() time.Time { return jobTestNow })
	if err != nil {
		t.Fatal(err)
	}
	result := provider.Execute(context.Background(), documentsEffectRequest(t, "upload"))
	receipt, ok := result.Receipt()
	if !ok || receipt.Code != "DOCUMENT_UPLOAD_RECONCILED" ||
		receipt.Status != effectsv1.ReceiptSucceeded {
		t.Fatalf("upload reconcile receipt=%+v", receipt)
	}
}

func TestAdoptionCopyProviderReturnsVerifiedReceipt(t *testing.T) {
	digest := sha256.Sum256([]byte("adoption ciphertext"))
	provider, err := newAdoptionCopyProvider(adoptionCopyEffectStub{
		checkpoint: adoptionapp.Checkpoint{RunID: jobTestResource,
			SourceVersionID: jobTestID, State: adoptionapp.CheckpointVerified,
			TargetObject: adoptionapp.ObjectReceipt{Version: "adoption-object-v1",
				CiphertextSHA256: digest}},
	}, func() time.Time { return jobTestNow })
	if err != nil {
		t.Fatal(err)
	}
	result := provider.Execute(context.Background(), documentsEffectRequest(t, "adoption"))
	receipt, ok := result.Receipt()
	if !ok || receipt.Code != "DOCUMENT_ADOPTION_COPY_VERIFIED" ||
		receipt.Status != effectsv1.ReceiptSucceeded {
		t.Fatalf("adoption copy receipt=%+v", receipt)
	}
}

func documentsEffectRequest(t *testing.T, kind string) effectsv1.Request {
	t.Helper()
	provider, operation, effectKind := "documents.archive", "build", "documents.archive.build"
	key := "documents:archive:" + jobTestResource.String()
	reference, digest, err := jobcontract.Archive(jobcontract.ID(jobTestTenant),
		jobcontract.ID(jobTestResource))
	cutoverGeneration := int64(1)
	if kind == "deletion" {
		provider, operation, effectKind = "documents.blob", "delete", "documents.blob.delete"
		key = "documents:blob-delete:" + jobTestResource.String()
		reference, digest, err = jobcontract.Deletion(jobcontract.ID(jobTestTenant),
			jobcontract.ID(jobTestResource))
	}
	if kind == "upload" {
		provider, operation, effectKind = "documents.upload", "reconcile",
			"documents.upload.reconcile"
		key = uploadReconcileKey(jobTestResource.String(), 1)
		reference, digest, err = jobcontract.UploadReconcile(jobcontract.ID(jobTestTenant),
			jobcontract.ID(jobTestResource), 1)
	}
	if kind == "adoption" {
		provider, operation, effectKind = "documents.adoption", "copy",
			"documents.adoption.copy"
		manifest := sha256.Sum256([]byte("documents adoption manifest"))
		cutoverGeneration = 17
		key = adoptionCopyKey(jobTestResource.String(), jobTestID.String())
		reference, digest, err = jobcontract.AdoptionCopy(jobcontract.ID(jobTestTenant),
			jobcontract.ID(jobTestResource), jobcontract.ID(jobTestID),
			uint64(cutoverGeneration), manifest)
	}
	if err != nil {
		t.Fatal(err)
	}
	tenant := jobTestTenant.String()
	request, err := effectsv1.NewRequest(
		"22222222-2222-4222-8222-222222222222", effectKind, 1, &tenant,
		effectsv1.Authority{ScopeKind: effectsv1.ScopeTenant, ScopeID: tenant,
			GroupID: "documents", GroupGeneration: 1, Domain: "documents",
			CutoverGeneration: cutoverGeneration, WriterOrigin: effectsv1.WriterV2},
		provider, operation, key, "documents-job-test", jobTestID.String(),
		reference, digest, jobTestNow.Format(time.RFC3339Nano))
	if err != nil {
		t.Fatal(err)
	}
	return request
}

func mustDocumentID(value string) foundation.ID {
	result, _ := foundation.ParseID(value)
	return result
}

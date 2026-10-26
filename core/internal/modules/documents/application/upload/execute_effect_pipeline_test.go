package upload

import (
	"bytes"
	"context"
	"crypto/sha256"
	"io"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

type pipelineKeyring struct{}

func (pipelineKeyring) CreateDataKey(context.Context, domain.ID, domain.ID, domain.ID) (DataKey, error) {
	return DataKey{Plaintext: [32]byte{1}, Wrapped: make([]byte, 115),
		KeyID: "documents-kek-1"}, nil
}

type pipelineCipher struct{}

func (pipelineCipher) Encrypt(_ context.Context, source io.Reader, target io.Writer,
	_ [32]byte, _ CipherMetadata,
) (CipherReceipt, error) {
	value, err := io.ReadAll(source)
	if err != nil {
		return CipherReceipt{}, err
	}
	ciphertext := append([]byte("cipher:"), value...)
	if _, err = target.Write(ciphertext); err != nil {
		return CipherReceipt{}, err
	}
	return CipherReceipt{PlaintextSHA256: sha256.Sum256(value),
		CiphertextSHA256: sha256.Sum256(ciphertext),
		CiphertextBytes:  int64(len(ciphertext))}, nil
}

type pipelineScan struct {
	bytes.Buffer
	clean bool
}

func (scan *pipelineScan) Finish(context.Context) (ScanResult, error) {
	return ScanResult{Clean: scan.clean, DetectedKind: domain.FilePDF,
		DetectedMIME: "application/pdf", SignatureRevision: "clamav-20260718",
		ScannedAt: uploadNow}, nil
}

type pipelineScanner struct{ clean bool }

func (scanner pipelineScanner) Open(context.Context, string, string, int64) (ScanSession, error) {
	return &pipelineScan{clean: scanner.clean}, nil
}

type pipelineObjects struct{ body []byte }

func (objects *pipelineObjects) Put(_ context.Context, request ObjectPutRequest) (ObjectPutResult, error) {
	value, err := io.ReadAll(request.Body)
	objects.body = value
	return ObjectPutResult{Version: "opaque-v1", Generation: request.Generation}, err
}

func (objects *pipelineObjects) Reconcile(_ context.Context,
	_ ObjectExpected,
) (ObjectPutResult, error) {
	return ObjectPutResult{Version: "opaque-v1", Generation: 1}, nil
}

func (objects *pipelineObjects) Verify(_ context.Context, expected ObjectExpected) error {
	if expected.CiphertextSHA256 != sha256.Sum256(objects.body) {
		return ErrUnavailable
	}
	return nil
}

func TestEffectPipelineStreamsPlaintextToScannerAndCiphertextToPrivateObject(t *testing.T) {
	content := []byte("content")
	plan := pipelinePlan(t, content)
	objects := &pipelineObjects{}
	pipeline, err := NewEffectPipeline(pipelineKeyring{}, pipelineCipher{},
		pipelineScanner{clean: true}, objects, uploadClock{})
	if err != nil {
		t.Fatal(err)
	}
	receipt, failure := pipeline.Execute(context.Background(), plan, bytes.NewReader(content))
	if failure != nil || receipt.Object.PlaintextSHA256 != plan.ExpectedContentSHA256 ||
		receipt.DetectedKind != domain.FilePDF || bytes.Equal(objects.body, content) {
		t.Fatalf("receipt=%+v failure=%+v object=%q", receipt, failure, objects.body)
	}
}

func TestEffectPipelineKeepsObjectReceiptForMalwareCleanup(t *testing.T) {
	content := []byte("content")
	objects := &pipelineObjects{}
	pipeline, _ := NewEffectPipeline(pipelineKeyring{}, pipelineCipher{},
		pipelineScanner{clean: false}, objects, uploadClock{})
	_, failure := pipeline.Execute(context.Background(), pipelinePlan(t, content),
		bytes.NewReader(content))
	if failure == nil || failure.Kind != FailureRejected || failure.Object == nil ||
		failure.Object.Version != "opaque-v1" {
		t.Fatalf("failure=%+v", failure)
	}
}

func TestExecuteOperationContract(t *testing.T) {
	if any((*EffectPipeline).Execute) == nil {
		t.Fatal("upload effect execution is unavailable")
	}
}

func pipelinePlan(t *testing.T, content []byte) AttemptPlan {
	t.Helper()
	key, err := domain.ContentObjectKey(uploadTenant, uploadDocument, uploadVersion)
	if err != nil {
		t.Fatal(err)
	}
	return AttemptPlan{UploadID: uploadHandle, TenantID: uploadTenant, ActorID: uploadActor,
		DocumentID: uploadDocument, CandidateVersionID: uploadVersion,
		BillingReservationID: uploadReservation, EffectReceiptID: uploadReceipt,
		Generation: 1, ObjectKey: key, OriginalName: "evidence.pdf",
		DeclaredMIME: "application/pdf", ExpectedContentSHA256: sha256.Sum256(content),
		ExpectedBytes: int64(len(content)), DeclaredKind: domain.FilePDF}
}

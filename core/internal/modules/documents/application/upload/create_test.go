package upload

import (
	"context"
	"crypto/sha256"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestCreateBoundReservesCiphertextAndPersistsDistinctIdentities(t *testing.T) {
	billing := &uploadBilling{}
	application := uploadApplication([]domain.ID{
		uploadHandle, uploadDocument, uploadVersion, uploadOperation, uploadAudit,
	}, billing)
	var captured CreateMutation
	store := uploadStore{create: func(_ context.Context, mutation CreateMutation) (Snapshot, error) {
		captured = mutation
		return mutation.Snapshot, nil
	}}
	request := CreateRequest{FolderID: uploadFolder, OriginalName: "evidence.pdf",
		DeclaredMIME: "application/pdf", IdempotencyKey: "upload-create-1",
		ExpectedContentSHA256: sha256.Sum256([]byte("content")), ExpectedBytes: 7}
	result, err := application.createBound(context.Background(), platformtx.TenantScope{},
		store, uploadIdentity(), request, sha256.Sum256([]byte("request")))
	if err != nil || result.ID != uploadHandle || result.CiphertextUpperBound != 117 ||
		captured.OperationID != uploadOperation || captured.AuditID != uploadAudit ||
		captured.BillingReservationID != uploadReservation {
		t.Fatalf("result=%+v mutation=%+v err=%v", result, captured, err)
	}
}

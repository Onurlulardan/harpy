package profilebackfill

import (
	"crypto/sha256"
	"testing"

	"github.com/google/uuid"
)

func TestReadContractsRejectUnboundedOrIncompleteInput(t *testing.T) {
	validID := uuid.MustParse("12000000-0000-4000-8000-000000000001")
	evidence := Evidence{
		OperationID: validID, RequestSHA256: sha256.Sum256([]byte("request")),
		ActorID: validID, ManagePermissionID: validID,
		ExpectedTokenVersion: 1, ExpectedAuthorizationVersion: 1,
	}
	if err := (ReadQuery{
		ResourceKind: "user", SourceSchemaID: validID, Limit: MaximumPageSize,
		Evidence: evidence,
	}).Validate(); err != nil {
		t.Fatalf("valid query rejected: %v", err)
	}
	if err := (ReadQuery{
		ResourceKind: "user", SourceSchemaID: validID, Limit: MaximumPageSize + 1,
		Evidence: evidence,
	}).Validate(); err == nil {
		t.Fatal("unbounded query passed")
	}
	if err := (SourceResource{
		ID: validID, SchemaVersionID: validID, Revision: 1, Document: []byte("{}"),
		SHA256: sha256.Sum256([]byte("{}")),
	}).Validate(); err != nil {
		t.Fatalf("valid source rejected: %v", err)
	}
}

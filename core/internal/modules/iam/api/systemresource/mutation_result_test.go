package systemresource

import (
	"crypto/sha256"
	"encoding/json"
	"testing"
)

func TestMutationResultRevisionContract(t *testing.T) {
	if _, err := mutationResultFixture(t, 4, 4, false); err != nil {
		t.Fatalf("current mutation rejected: %v", err)
	}
	if _, err := mutationResultFixture(t, 6, 4, true); err != nil {
		t.Fatalf("replay with newer record rejected: %v", err)
	}
	if _, err := mutationResultFixture(t, 6, 4, false); err == nil {
		t.Fatal("non-replay accepted a different record revision")
	}
	if _, err := mutationResultFixture(t, 3, 4, true); err == nil {
		t.Fatal("replay accepted a record older than its receipt")
	}
}

func mutationResultFixture(
	t *testing.T,
	recordRevision int64,
	receiptRevision int64,
	replayed bool,
) (MutationResult, error) {
	t.Helper()
	fieldID := mustSystemResourceID(t, UserIDFieldID)
	catalog := testUserFieldCatalog(t)
	recordID := mustSystemResourceID(t, "018f0f4e-7b61-7c25-8000-000000000020")
	values, err := NewFieldValues(map[ID]json.RawMessage{
		fieldID: json.RawMessage(`"018f0f4e-7b61-7c25-8000-000000000020"`),
	})
	if err != nil {
		t.Fatal(err)
	}
	record, err := NewRecord(Users, recordID, recordRevision, values)
	if err != nil {
		t.Fatal(err)
	}
	read, err := NewReadResult(record, catalog)
	if err != nil {
		t.Fatal(err)
	}
	receipt, err := NewMutationReceipt(
		mustSystemResourceID(t, "018f0f4e-7b61-7c25-8000-000000000021"),
		sha256.Sum256([]byte("request")), receiptRevision, replayed,
	)
	if err != nil {
		t.Fatal(err)
	}
	return NewMutationResult(read, receipt)
}

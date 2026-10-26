package systemresource

import (
	"crypto/sha256"
	"encoding/json"
	"testing"
)

func TestRecordAndReceiptRequireCanonicalIdentity(t *testing.T) {
	fieldID := mustSystemResourceID(t, "a1000000-0000-4000-8000-000000000001")
	values, _ := NewFieldValues(map[ID]json.RawMessage{
		fieldID: []byte(`"018f0f4e-7b61-7c25-8000-000000000020"`),
	})
	recordID := mustSystemResourceID(t, "018f0f4e-7b61-7c25-8000-000000000020")
	record, err := NewRecord(Users, recordID, 4, values)
	if err != nil || record.Validate() != nil {
		t.Fatalf("record error=%v", err)
	}
	receipt, err := NewMutationReceipt(
		mustSystemResourceID(t, "018f0f4e-7b61-7c25-8000-000000000021"),
		sha256.Sum256([]byte("request")), 4, false,
	)
	revision, replayed, resultErr := receipt.Result()
	if err != nil || resultErr != nil || revision != 4 || replayed {
		t.Fatalf("receipt revision=%d replayed=%v error=%v", revision, replayed, err)
	}
}

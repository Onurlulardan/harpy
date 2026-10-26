package httpapi

import (
	"encoding/json"
	"testing"
)

func TestQueryViewsPreserveCanonicalSparseRecords(t *testing.T) {
	raw := json.RawMessage(`{"id":"019f0000-0000-7000-8000-000000000001"}`)
	values := []any{NewSubscriptionQueryView(raw), NewAuditEventQueryView(raw)}
	for _, value := range values {
		encoded, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		if string(encoded) != string(raw) {
			t.Fatalf("query view = %s, want %s", encoded, raw)
		}
	}
}

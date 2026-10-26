package runtimev1

import (
	"encoding/json"
	"testing"
)

func typedContractFixture[T any](t *testing.T, value map[string]any) T {
	t.Helper()
	document, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal typed contract fixture: %v", err)
	}
	var decoded T
	if err := json.Unmarshal(document, &decoded); err != nil {
		t.Fatalf("decode typed contract fixture: %v", err)
	}
	return decoded
}

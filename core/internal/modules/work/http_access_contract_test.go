package work

import (
	"testing"

	"github.com/google/uuid"
)

func TestHTTPAccessContractRejectsAuthorityResourceAndEdgeDrift(t *testing.T) {
	access, err := NewHTTPAccess(uuid.New(), uuid.New(), 1, 1,
		"work.task.get", "getWorkTask", "", "", false)
	if err != nil {
		t.Fatal(err)
	}
	if !validHTTPAccess(access, "getWorkTask", "tasks", false) {
		t.Fatal("exact access contract rejected")
	}
	for _, mismatch := range []struct {
		operation, resource string
		compatibility       bool
	}{
		{"getWorkBoard", "tasks", false},
		{"getWorkTask", "boards", false},
		{"getWorkTask", "tasks", true},
	} {
		if validHTTPAccess(access, mismatch.operation, mismatch.resource,
			mismatch.compatibility) {
			t.Fatalf("mismatched contract accepted: %#v", mismatch)
		}
	}
}

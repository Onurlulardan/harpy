package permissionrequest

import (
	"testing"

	"github.com/google/uuid"
)

func TestReadQueryRequiresGroupIdentity(t *testing.T) {
	if err := (ReadQuery{}).Validate(); err == nil {
		t.Fatal("empty group identity was accepted")
	}
	if err := (ReadQuery{GroupID: uuid.New()}).Validate(); err != nil {
		t.Fatalf("valid read query: %v", err)
	}
}

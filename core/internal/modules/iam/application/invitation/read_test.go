package invitation

import (
	"testing"

	"github.com/google/uuid"
)

func TestReadQueryRequiresInvitationIdentity(t *testing.T) {
	if err := (ReadQuery{}).Validate(); err == nil {
		t.Fatal("empty invitation identity was accepted")
	}
	if err := (ReadQuery{ID: uuid.New()}).Validate(); err != nil {
		t.Fatalf("valid read query: %v", err)
	}
}

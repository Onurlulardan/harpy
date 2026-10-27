package profilepostgres

import (
	"testing"
)

func TestPrepareInvitationUserProfileRejectsMissingCatalog(t *testing.T) {
	if _, err := PrepareInvitationUserProfile(
		nil, "Ada", "Lovelace", nil,
	); err == nil {
		t.Fatal("missing profile catalog was accepted")
	}
}

package iam

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestInvitationHTTPViewsCannotSerializeCredentialMaterial(t *testing.T) {
	views := []any{
		InvitationView{
			ID: uuid.New(), Email: "invitee@example.com", Role: "Operator",
			InvitedBy: uuid.New(), Status: "PENDING", ExpiresAt: time.Now(),
			CreatedAt: time.Now(), Revision: 1,
		},
		InvitationVerifyView{
			Email: "invitee@example.com", Role: "Operator",
			TenantName: "acme", TenantSlug: "acme", ExpiresAt: time.Now(),
		},
		InvitationAcceptView{
			UserID: uuid.New(), TenantSlug: "acme", RequiresLogin: true,
		},
	}
	for _, view := range views {
		document, err := json.Marshal(view)
		if err != nil {
			t.Fatal(err)
		}
		lower := bytes.ToLower(document)
		for _, forbidden := range [][]byte{
			[]byte("token"), []byte("secret"), []byte("hmac"),
			[]byte("ciphertext"), []byte("artifact"), []byte("session"),
		} {
			if bytes.Contains(lower, forbidden) {
				t.Fatalf("response exposed forbidden material %q: %s", forbidden, document)
			}
		}
	}
}

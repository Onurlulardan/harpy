package iam

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"

	rbacapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/rbacchange"
)

func TestRBACEventPayloadIsCanonicalAndRedacted(t *testing.T) {
	change := rbacapp.Change{
		OperationID: [16]byte{1}, ActorID: [16]byte{2}, AggregateID: [16]byte{3},
		RequestSHA256: [32]byte{4}, CommandKey: "iam.rbac.resource.update",
		ResultCode: "iam.rbac.resource.updated", Revision: 7, Clock: 11,
	}
	payload, digest, err := rbacEventPayload(change)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"clock":11,"command":"iam.rbac.resource.update","result":"iam.rbac.resource.updated","revision":7}`
	if string(payload) != want {
		t.Fatalf("RBAC payload = %s", payload)
	}
	for _, forbidden := range []string{
		change.ActorID.String(), change.AggregateID.String(),
		hex.EncodeToString(change.RequestSHA256[:]),
	} {
		if strings.Contains(string(payload), forbidden) {
			t.Fatalf("RBAC payload leaked %q", forbidden)
		}
	}
	hash := sha256.Sum256(payload)
	if digest != hex.EncodeToString(hash[:]) {
		t.Fatal("RBAC payload digest is not canonical payload SHA-256")
	}
}

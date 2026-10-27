package iam

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/invitation"
	"github.com/google/uuid"
)

func TestInvitationEventPayloadIsCanonicalAndSecretFree(t *testing.T) {
	change := invitation.AuditChange{
		OperationID: uuid.New(), ActorID: uuid.New(), InvitationID: uuid.New(),
		RequestSHA256: [32]byte{7}, CommandKey: "iam.invitation.create",
		ResultState: invitation.StatePending, ResultRevision: 1,
		DeliveryArtifactRef: "iam-delivery://tenant/aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa/" +
			"artifact/bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb",
	}
	payload, digest, err := invitationEventPayload(change)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"command":"iam.invitation.create","delivery_artifact_ref":"` +
		change.DeliveryArtifactRef + `","revision":1,"state":"pending"}`
	if string(payload) != want {
		t.Fatalf("payload = %s", payload)
	}
	for _, forbidden := range []string{
		change.ActorID.String(), change.InvitationID.String(),
		hex.EncodeToString(change.RequestSHA256[:]), "token", "secret", "hmac",
	} {
		if strings.Contains(string(payload), forbidden) {
			t.Fatalf("payload leaked %q", forbidden)
		}
	}
	hash := sha256.Sum256(payload)
	if digest != hex.EncodeToString(hash[:]) {
		t.Fatal("payload digest mismatch")
	}
}

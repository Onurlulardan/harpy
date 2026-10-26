package iam

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"

	changeapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/change"
	"github.com/google/uuid"
)

func TestIdentityChangePayloadIsCanonicalAndRedacted(t *testing.T) {
	evidence := identityChangeEvidence()
	aggregateID := uuid.MustParse("ab000000-0000-4000-8000-000000000001")
	change := changeapp.NewChange(
		evidence, [16]byte(aggregateID), "user", "iam.identity.user.create",
	)
	record := changeapp.RecordedChange{
		Change: change, Revision: 5, Clock: 11, EventSequence: 1,
	}
	payload, digest, err := identityChangeEventPayload(record)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"clock":11,"command":"iam.identity.user.create","resource":"user","result":"iam.identity.user.create.succeeded","revision":5}`
	if string(payload) != want {
		t.Fatalf("identity payload = %s", payload)
	}
	for _, forbidden := range []string{
		uuid.UUID(change.ActorID).String(), aggregateID.String(),
		hex.EncodeToString(change.RequestSHA256[:]), "password", "profile",
		"user@example.com",
	} {
		if strings.Contains(string(payload), forbidden) {
			t.Fatalf("identity payload leaked %q", forbidden)
		}
	}
	hash := sha256.Sum256(payload)
	if digest != hex.EncodeToString(hash[:]) {
		t.Fatal("identity payload digest is not canonical payload SHA-256")
	}
}

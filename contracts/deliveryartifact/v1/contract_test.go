package deliveryartifactv1

import (
	"testing"

	"github.com/google/uuid"
)

const fixtureKey = "gcp-kms://projects/harpy-prod/locations/europe-west1/" +
	"keyRings/iam/cryptoKeys/delivery/cryptoKeyVersions/7"

func TestReferenceIsCanonicalAndTenantBound(t *testing.T) {
	tenantID, artifactID := uuid.New(), uuid.New()
	reference, err := NewReference(tenantID, artifactID)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := ParseReference(reference.String())
	if err != nil || parsed != reference {
		t.Fatalf("round trip = %#v, %v", parsed, err)
	}
	for _, invalid := range []string{
		"", reference.String() + "/secret", stringsUpper(reference.String()),
		"kms://invitation/artifact-1",
	} {
		if _, err := ParseReference(invalid); err == nil {
			t.Fatalf("invalid reference accepted: %q", invalid)
		}
	}
}

func TestEnvelopeBindsEveryOwnerDimension(t *testing.T) {
	tenantID, artifactID, invitationID, tokenID :=
		uuid.New(), uuid.New(), uuid.New(), uuid.New()
	aad, _ := AADSHA256(tenantID, artifactID, invitationID, tokenID, fixtureKey)
	envelope := NewEnvelope(
		artifactID, invitationID, tokenID, fixtureKey,
		[NonceBytes]byte{1}, [CiphertextBytes]byte{2}, [TagBytes]byte{3}, aad,
	)
	if err := envelope.Validate(tenantID); err != nil {
		t.Fatal(err)
	}
	for name, mutate := range map[string]func(*Envelope){
		"artifact":   func(value *Envelope) { value.ArtifactID = uuid.New() },
		"invitation": func(value *Envelope) { value.InvitationID = uuid.New() },
		"token":      func(value *Envelope) { value.TokenID = uuid.New() },
		"ciphertext": func(value *Envelope) { value.Ciphertext[0] ^= 1 },
		"tag":        func(value *Envelope) { value.Tag[0] ^= 1 },
	} {
		t.Run(name, func(t *testing.T) {
			changed := envelope
			mutate(&changed)
			if changed.Validate(tenantID) == nil {
				t.Fatal("substitution accepted")
			}
		})
	}
	if envelope.Validate(uuid.New()) == nil {
		t.Fatal("cross-tenant envelope accepted")
	}
}

func stringsUpper(value string) string {
	bytes := []byte(value)
	for index, current := range bytes {
		if current >= 'a' && current <= 'z' {
			bytes[index] = current - ('a' - 'A')
			break
		}
	}
	return string(bytes)
}

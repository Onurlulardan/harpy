package lifecycledeliveryv1

import (
	"testing"

	"github.com/google/uuid"
)

func TestLifecycleEnvelopeAndReferenceArePurposeBound(t *testing.T) {
	tenant := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	artifact := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	subject := uuid.MustParse("33333333-3333-4333-8333-333333333333")
	resource := uuid.MustParse("44444444-4444-4444-8444-444444444444")
	token := uuid.MustParse("55555555-5555-4555-8555-555555555555")
	key := "gcp-kms://projects/harpy-prod/locations/europe-west1/keyRings/iam/cryptoKeys/delivery/cryptoKeyVersions/7"
	aad, err := AADSHA256(tenant, artifact, subject, resource, token,
		PurposePasswordReset, key)
	if err != nil {
		t.Fatal(err)
	}
	var nonce [NonceBytes]byte
	var ciphertext [CiphertextBytes]byte
	var tag [TagBytes]byte
	nonce[0], ciphertext[0], tag[0] = 1, 2, 3
	envelope := NewEnvelope(artifact, subject, resource, token,
		PurposePasswordReset, key, nonce, ciphertext, tag, aad)
	if err := envelope.Validate(tenant); err != nil {
		t.Fatal(err)
	}
	envelope.Purpose = PurposePortalSetup
	if envelope.Validate(tenant) == nil {
		t.Fatal("cross-purpose envelope substitution was accepted")
	}
	reference, err := NewReference(tenant, artifact)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := ParseReference(reference.String())
	if err != nil || parsed.TenantID() != tenant || parsed.ArtifactID() != artifact {
		t.Fatal("lifecycle delivery reference did not round-trip")
	}
}

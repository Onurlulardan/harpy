package licenseauthority

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"testing"
	"time"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	"github.com/google/uuid"
)

type authorityFixture struct {
	now        time.Time
	binding    Binding
	privateKey ed25519.PrivateKey
	publicKey  ed25519.PublicKey
	payload    wirePayload
}

func newAuthorityFixture(t *testing.T) authorityFixture {
	t.Helper()
	publicKey, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)
	binding := Binding{
		LicenseID: uuid.New(), LicenseeOrganizationID: uuid.New(),
		EnvironmentID: uuid.New(), DeliveryKind: DeliveryConnected,
	}
	return authorityFixture{
		now: now, binding: binding, privateKey: privateKey, publicKey: publicKey,
		payload: wirePayload{
			Schema: SnapshotSchema, Issuer: "harpy", Audience: SnapshotAudience,
			Product: SnapshotProduct, LicenseID: binding.LicenseID.String(),
			LicenseeOrganizationID: binding.LicenseeOrganizationID.String(),
			EnvironmentID:          binding.EnvironmentID.String(), DeploymentMode: "onpremise",
			DeliveryKind: binding.DeliveryKind, Sequence: 7, Status: StatusActive,
			IssuedAt:            now.Add(-2 * time.Hour).Format(time.RFC3339),
			NotBefore:           now.Add(-time.Hour).Format(time.RFC3339),
			RefreshAfter:        now.Add(time.Hour).Format(time.RFC3339),
			GrantValidUntil:     now.Add(2 * time.Hour).Format(time.RFC3339),
			GraceUntil:          now.Add(3 * time.Hour).Format(time.RFC3339),
			CommercialExpiresAt: now.Add(24 * time.Hour).Format(time.RFC3339),
			Features:            wireFeatures{FormBuilder: true, APIBuilder: true},
			Limits:              wireLimits{MaxUsers: 25, MaxForms: 0, MaxTenants: 1},
			ReleaseChannel:      "stable",
		},
	}
}

func (fixture authorityFixture) document(t *testing.T) []byte {
	t.Helper()
	payload, err := json.Marshal(fixture.payload)
	if err != nil {
		t.Fatal(err)
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(payload, MaximumDocumentBytes)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(canonical)
	envelope := wireEnvelope{
		Schema: EnvelopeSchema, KeyID: "license-root-1", Algorithm: SignatureAlgorithm,
		Canonicalization: Canonicalization, PayloadSHA256: hex.EncodeToString(digest[:]),
		Payload: canonical,
		Signature: base64.RawURLEncoding.EncodeToString(
			ed25519.Sign(fixture.privateKey, signatureMessage(canonical)),
		),
	}
	document, err := json.Marshal(envelope)
	if err != nil {
		t.Fatal(err)
	}
	document, err = canonicaljsonv1.CanonicalizeObject(document, MaximumDocumentBytes)
	if err != nil {
		t.Fatal(err)
	}
	return document
}

func (fixture authorityFixture) resolver() KeyResolver {
	return KeyResolverFunc(func(
		context.Context, string, string, time.Time, time.Time,
	) (ed25519.PublicKey, error) {
		return append(ed25519.PublicKey(nil), fixture.publicKey...), nil
	})
}

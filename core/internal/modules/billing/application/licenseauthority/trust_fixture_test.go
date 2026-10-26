package licenseauthority

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"testing"
	"time"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
)

type trustFixture struct {
	now                        time.Time
	rootPublic, signerPublic   ed25519.PublicKey
	rootPrivate, signerPrivate ed25519.PrivateKey
	rootStatus, signerStatus   TrustKeyStatus
	issuer                     string
}

func newTrustFixture(t *testing.T) trustFixture {
	t.Helper()
	rootPublic, rootPrivate, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	signerPublic, signerPrivate, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	return trustFixture{
		now:        time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC),
		rootPublic: rootPublic, rootPrivate: rootPrivate,
		signerPublic: signerPublic, signerPrivate: signerPrivate,
		rootStatus: TrustKeyActive, signerStatus: TrustKeyActive, issuer: "forgeguard",
	}
}

func (fixture trustFixture) document(t *testing.T) []byte {
	t.Helper()
	instant := func(offset time.Duration) string {
		return fixture.now.Add(offset).Format(time.RFC3339)
	}
	signer := trustSignerWire{
		KeyID: "license-root-1", Algorithm: SignatureAlgorithm,
		PublicKey: base64.RawURLEncoding.EncodeToString(fixture.signerPublic),
		Status:    fixture.signerStatus, NotBefore: instant(-48 * time.Hour),
		SignUntil: instant(24 * time.Hour), VerifyUntil: instant(72 * time.Hour),
	}
	payload := keysetPayloadWire{
		Schema: KeysetSchema, Issuer: fixture.issuer, Audience: KeysetAudience,
		Sequence: 4, IssuedAt: instant(-2 * time.Hour), NotBefore: instant(-time.Hour),
		ExpiresAt: instant(48 * time.Hour), SnapshotSigners: []trustSignerWire{signer},
	}
	canonicalPayload := canonicalTestDocument(t, payload)
	digest := sha256.Sum256(canonicalPayload)
	envelope := keysetEnvelopeWire{
		Schema: KeysetEnvelopeSchema, KeyID: "release-1", Algorithm: SignatureAlgorithm,
		Canonicalization: Canonicalization, PayloadSHA256: hex.EncodeToString(digest[:]),
		Payload: canonicalPayload,
		Signature: base64.RawURLEncoding.EncodeToString(
			ed25519.Sign(fixture.rootPrivate, keysetSignatureMessage(canonicalPayload))),
	}
	root := trustRootWire{
		KeyID: "release-1", Algorithm: SignatureAlgorithm,
		PublicKey: base64.RawURLEncoding.EncodeToString(fixture.rootPublic),
		Status:    fixture.rootStatus, NotBefore: instant(-72 * time.Hour),
		SignUntil: instant(24 * time.Hour), VerifyUntil: instant(96 * time.Hour),
	}
	return canonicalTestDocument(t, trustBundleWire{
		Schema: TrustBundleSchema, Purpose: EnvelopeSchema,
		Roots: []trustRootWire{root}, Keysets: []keysetEnvelopeWire{envelope},
	})
}

func canonicalTestDocument(t *testing.T, value any) []byte {
	t.Helper()
	document, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(document, MaximumDocumentBytes)
	if err != nil {
		t.Fatal(err)
	}
	return canonical
}

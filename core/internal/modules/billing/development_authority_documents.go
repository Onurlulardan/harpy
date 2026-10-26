package billing

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"time"

	installationv1 "github.com/Onurlulardan/harpy/contracts/installation/v1"
	licensev2 "github.com/Onurlulardan/harpy/contracts/license/v2"
	"github.com/Onurlulardan/harpy/core/internal/platform/localinstall"
	"github.com/google/uuid"
)

type developmentAuthorityDocuments struct {
	rootSeed, signerSeed []byte
	trust, snapshot      []byte
}

func (documents *developmentAuthorityDocuments) destroy() {
	clear(documents.rootSeed)
	clear(documents.signerSeed)
	clear(documents.trust)
	clear(documents.snapshot)
}

func buildDevelopmentAuthorityDocuments(
	manifest installationv1.Manifest,
	now time.Time,
) (*developmentAuthorityDocuments, error) {
	rootPublic, rootPrivate, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	defer clear(rootPrivate)
	signerPublic, signerPrivate, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	defer clear(signerPrivate)
	trust, err := buildDevelopmentTrustBundle(now, rootPublic, rootPrivate, signerPublic)
	if err != nil {
		return nil, err
	}
	snapshot, err := buildDevelopmentSnapshot(manifest, now, signerPrivate)
	if err != nil {
		clear(trust)
		return nil, err
	}
	return &developmentAuthorityDocuments{
		rootSeed:   append([]byte(nil), rootPrivate.Seed()...),
		signerSeed: append([]byte(nil), signerPrivate.Seed()...),
		trust:      trust, snapshot: snapshot,
	}, nil
}

func buildDevelopmentTrustBundle(
	now time.Time,
	rootPublic ed25519.PublicKey,
	rootPrivate ed25519.PrivateKey,
	signerPublic ed25519.PublicKey,
) ([]byte, error) {
	instant := func(offset time.Duration) string {
		return now.Add(offset).UTC().Format(time.RFC3339)
	}
	signer := licensev2.SnapshotSigner{
		KeyID: developmentSignerKeyID, Algorithm: licensev2.SignatureAlgorithm,
		PublicKey: base64.RawURLEncoding.EncodeToString(signerPublic),
		Status:    licensev2.TrustKeyActive, NotBefore: instant(-time.Hour),
		SignUntil:   instant(10 * 365 * 24 * time.Hour),
		VerifyUntil: instant(11 * 365 * 24 * time.Hour),
	}
	payload := licensev2.KeysetPayload{
		Schema: licensev2.KeysetSchema, Issuer: developmentLicenseIssuer,
		Audience: licensev2.KeysetAudience, Sequence: 1,
		IssuedAt: instant(-2 * time.Minute), NotBefore: instant(-time.Minute),
		ExpiresAt:       instant(10 * 365 * 24 * time.Hour),
		SnapshotSigners: []licensev2.SnapshotSigner{signer},
	}
	envelope, err := signDevelopmentKeyset(payload, rootPrivate)
	if err != nil {
		return nil, err
	}
	root := licensev2.TrustRoot{
		KeyID: developmentRootKeyID, Algorithm: licensev2.SignatureAlgorithm,
		PublicKey: base64.RawURLEncoding.EncodeToString(rootPublic),
		Status:    licensev2.TrustKeyActive, NotBefore: instant(-time.Hour),
		SignUntil:   instant(10 * 365 * 24 * time.Hour),
		VerifyUntil: instant(11 * 365 * 24 * time.Hour),
	}
	return canonicalDevelopmentLicense(licensev2.TrustBundle{
		Schema: licensev2.TrustBundleSchema, Purpose: licensev2.SnapshotEnvelopeSchema,
		Roots: []licensev2.TrustRoot{root}, Keysets: []licensev2.KeysetEnvelope{envelope},
	})
}

func buildDevelopmentSnapshot(
	manifest installationv1.Manifest,
	now time.Time,
	signer ed25519.PrivateKey,
) ([]byte, error) {
	environmentID := uuid.UUID(manifest.EnvironmentID())
	identity := localinstall.DeriveIdentity(environmentID)
	instant := func(offset time.Duration) string {
		return now.Add(offset).UTC().Format(time.RFC3339)
	}
	payload := licensev2.SnapshotPayload{
		Schema: licensev2.SnapshotSchema, Issuer: developmentLicenseIssuer,
		Audience: licensev2.SnapshotAudience, Product: licensev2.SnapshotProduct,
		LicenseID:              identity.LicenseID.String(),
		LicenseeOrganizationID: identity.LicenseeOrganizationID.String(),
		EnvironmentID:          environmentID.String(), DeploymentMode: "onpremise",
		DeliveryKind: licensev2.DeliveryDevelopment, Sequence: 1,
		Status: licensev2.StatusActive, IssuedAt: instant(-2 * time.Minute),
		NotBefore: instant(-time.Minute), RefreshAfter: instant(365 * 24 * time.Hour),
		GrantValidUntil:     instant(10 * 365 * 24 * time.Hour),
		GraceUntil:          instant(10*365*24*time.Hour + 24*time.Hour),
		CommercialExpiresAt: instant(10 * 365 * 24 * time.Hour),
		Features: licensev2.Features{FormBuilder: true, WorkflowBuilder: true,
			APIBuilder: true, DashboardBuilder: true, AIAssistant: true},
		Limits: licensev2.Limits{}, ReleaseChannel: "stable",
	}
	return signDevelopmentSnapshot(payload, signer)
}

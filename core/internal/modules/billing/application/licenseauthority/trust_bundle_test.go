package licenseauthority

import (
	"context"
	"crypto/ed25519"
	"errors"
	"testing"
	"time"
)

func TestTrustBundleBuildsIssuerBoundResolver(t *testing.T) {
	fixture := newTrustFixture(t)
	bundle, err := ParseTrustBundle(context.Background(), fixture.document(t), fixture.now)
	if err != nil {
		t.Fatal(err)
	}
	if len(bundle.DigestHex()) != 64 {
		t.Fatalf("missing bundle digest: %q", bundle.DigestHex())
	}
	key, err := bundle.LicenseKey(
		context.Background(), fixture.issuer, "license-root-1",
		fixture.now.Add(-30*time.Minute), fixture.now,
	)
	if err != nil || !key.Equal(fixture.signerPublic) {
		t.Fatalf("trusted signer not resolved: %v", err)
	}
	key[0] ^= 1
	again, err := bundle.LicenseKey(
		context.Background(), fixture.issuer, "license-root-1",
		fixture.now.Add(-30*time.Minute), fixture.now,
	)
	if err != nil || again[0] != fixture.signerPublic[0] {
		t.Fatal("resolver leaked mutable key bytes")
	}
}

func TestTrustBundleRejectsWrongIssuerAndRevokedSigner(t *testing.T) {
	fixture := newTrustFixture(t)
	bundle, err := ParseTrustBundle(context.Background(), fixture.document(t), fixture.now)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = bundle.LicenseKey(
		context.Background(), "other", "license-root-1", fixture.now, fixture.now,
	); !errors.Is(err, ErrUntrustedKey) {
		t.Fatalf("wrong issuer resolved: %v", err)
	}
	fixture.signerStatus = TrustKeyRevoked
	bundle, err = ParseTrustBundle(context.Background(), fixture.document(t), fixture.now)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = bundle.LicenseKey(
		context.Background(), fixture.issuer, "license-root-1", fixture.now, fixture.now,
	); !errors.Is(err, ErrUntrustedKey) {
		t.Fatalf("revoked signer resolved: %v", err)
	}
}

func TestTrustBundleVerifiesSnapshotThroughPurposeSeparatedChain(t *testing.T) {
	trust := newTrustFixture(t)
	bundle, err := ParseTrustBundle(context.Background(), trust.document(t), trust.now)
	if err != nil {
		t.Fatal(err)
	}
	snapshot := newAuthorityFixture(t)
	snapshot.now = trust.now
	snapshot.privateKey = trust.signerPrivate
	snapshot.publicKey = trust.signerPublic
	snapshot.payload.Issuer = trust.issuer
	verified, err := Verify(
		context.Background(), snapshot.document(t), snapshot.binding, snapshot.now, bundle,
	)
	if err != nil || verified.Sequence() != snapshot.payload.Sequence {
		t.Fatalf("release-root to keyset to snapshot chain failed: %v", err)
	}
	trust.rootStatus = TrustKeyRevoked
	if _, err = ParseTrustBundle(
		context.Background(), trust.document(t), trust.now,
	); !errors.Is(err, ErrUntrustedTrustRoot) {
		t.Fatalf("revoked release root accepted: %v", err)
	}
}

func TestTrustBundleRejectsTamperNonCanonicalAndCrossPurposeKey(t *testing.T) {
	fixture := newTrustFixture(t)
	document := fixture.document(t)
	document[len(document)-8] ^= 1
	if _, err := ParseTrustBundle(context.Background(), document, fixture.now); err == nil {
		t.Fatal("tampered trust bundle accepted")
	}
	nonCanonical := append([]byte(" \n"), fixture.document(t)...)
	if _, err := ParseTrustBundle(context.Background(), nonCanonical, fixture.now); err == nil {
		t.Fatal("non-canonical trust bundle accepted")
	}
	fixture.signerPublic = append(ed25519.PublicKey(nil), fixture.rootPublic...)
	if _, err := ParseTrustBundle(
		context.Background(), fixture.document(t), fixture.now,
	); !errors.Is(err, ErrInvalidTrustBundle) {
		t.Fatalf("cross-purpose key reuse accepted: %v", err)
	}
}

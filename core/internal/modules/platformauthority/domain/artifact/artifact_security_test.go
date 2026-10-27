package artifact

import (
	"bytes"
	"encoding/base64"
	"testing"
	"time"
)

func TestArtifactsRejectUnknownDuplicateAndTamper(t *testing.T) {
	fixture := newArtifactFixture(t)
	unknown := bytes.Replace(
		fixture.proposal.Bytes, []byte(`"deployment_mode"`),
		[]byte(`"unexpected":true,"deployment_mode"`), 1,
	)
	duplicate := bytes.Replace(
		fixture.proposal.Bytes, []byte(`"schema"`),
		[]byte(`"schema":"protected-authority-proposal.v1","schema"`), 1,
	)
	tampered := bytes.Replace(
		fixture.proposal.Bytes, []byte(fixtureOperationID),
		[]byte("99999999-9999-4999-8999-999999999999"), 1,
	)
	for _, document := range [][]byte{unknown, duplicate, tampered} {
		if _, err := VerifyProposal(
			document, fixture.state, ExpiryOnline, fixtureNow, fixture.proposalPublic,
		); err == nil {
			t.Fatal("ambiguous or tampered proposal was accepted")
		}
	}
}

func TestTrustBundleRejectsSelfSignTOFUWrongKindAndRevoked(t *testing.T) {
	fixture := newArtifactFixture(t)
	selfBytes, err := SignTrustBundle(TrustBundle{
		Schema: TrustBundleSchema, RegistryKind: RegistryDevelopment, TrustGeneration: 1,
		Active: TrustKey{
			KID: "self", Algorithm: AlgorithmEd25519,
			PublicKey: base64.RawURLEncoding.EncodeToString(fixture.releasePublic),
			NotBefore: fixtureNow.Add(-time.Hour).Format(time.RFC3339),
			SignUntil: fixtureNow.Add(time.Hour).Format(time.RFC3339),
		}, Retired: []RetiredTrustKey{},
	}, fixture.releasePrivate)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := VerifyTrustBundle(
		selfBytes, RegistryDevelopment, fixtureNow, fixture.releasePublic,
	); err == nil {
		t.Fatal("self-signed registry root was accepted")
	}
	if _, err := VerifyTrustBundle(
		mustTrustBytes(t, fixture, nil), RegistrySaaS, fixtureNow, fixture.releasePublic,
	); err == nil {
		t.Fatal("wrong registry kind was accepted")
	}
	revoked := RetiredTrustKey{
		KID: "registry.v1", Algorithm: AlgorithmEd25519,
		PublicKey:   base64.RawURLEncoding.EncodeToString(fixture.registryPublic),
		SignUntil:   fixtureNow.Format(time.RFC3339),
		VerifyUntil: fixtureNow.Add(time.Hour).Format(time.RFC3339), Disposition: "revoked",
	}
	rotated := mustTrustBundle(t, fixture, []RetiredTrustKey{revoked})
	if _, err := VerifyAck(fixture.ack(t, fixture.registryPrivate).Bytes,
		fixture.proposal, rotated, fixtureNow); err == nil {
		t.Fatal("revoked registry KID was accepted")
	}
}

func TestCrossPurposeKeyCannotIssueRegistryAck(t *testing.T) {
	fixture := newArtifactFixture(t)
	ack := fixture.ack(t, fixture.proposalPrivate)
	if _, err := VerifyAck(ack.Bytes, fixture.proposal, fixture.trust, fixtureNow); err == nil {
		t.Fatal("proposal key issued a registry acknowledgement")
	}
}

func TestRetiredKeyVerifiesHistoryUntilWindowEnds(t *testing.T) {
	fixture := newArtifactFixture(t)
	retired := RetiredTrustKey{
		KID: "registry.v1", Algorithm: AlgorithmEd25519,
		PublicKey:   base64.RawURLEncoding.EncodeToString(fixture.registryPublic),
		SignUntil:   fixtureNow.Add(time.Minute).Format(time.RFC3339),
		VerifyUntil: fixtureNow.Add(time.Hour).Format(time.RFC3339),
		Disposition: "verify_only",
	}
	rotated := mustTrustBundle(t, fixture, []RetiredTrustKey{retired})
	ack := fixture.ack(t, fixture.registryPrivate)
	if _, err := VerifyAck(
		ack.Bytes, fixture.proposal, rotated, fixtureNow.Add(30*time.Minute),
	); err != nil {
		t.Fatalf("historical ack lost during rotation: %v", err)
	}
	if _, err := VerifyAck(
		ack.Bytes, fixture.proposal, rotated, fixtureNow.Add(2*time.Hour),
	); err == nil {
		t.Fatal("retired key verified beyond its window")
	}
}

func mustTrustBundle(
	t *testing.T, fixture artifactFixture, retired []RetiredTrustKey,
) VerifiedTrustBundle {
	t.Helper()
	activePublic, _ := deterministicKey(9)
	bytes, err := SignTrustBundle(TrustBundle{
		Schema: TrustBundleSchema, RegistryKind: RegistryDevelopment, TrustGeneration: 2,
		Active: TrustKey{KID: "registry.v2", Algorithm: AlgorithmEd25519,
			PublicKey: base64.RawURLEncoding.EncodeToString(activePublic),
			NotBefore: fixtureNow.Add(-time.Minute).Format(time.RFC3339),
			SignUntil: fixtureNow.Add(time.Hour).Format(time.RFC3339)}, Retired: retired,
	}, fixture.releasePrivate)
	if err != nil {
		t.Fatal(err)
	}
	bundle, err := VerifyTrustBundle(bytes, RegistryDevelopment, fixtureNow, fixture.releasePublic)
	if err != nil {
		t.Fatal(err)
	}
	return bundle
}

func mustTrustBytes(
	t *testing.T, fixture artifactFixture, retired []RetiredTrustKey,
) []byte {
	t.Helper()
	bytes, err := SignTrustBundle(TrustBundle{
		Schema: TrustBundleSchema, RegistryKind: RegistryDevelopment, TrustGeneration: 1,
		Active: TrustKey{KID: "registry.v1", Algorithm: AlgorithmEd25519,
			PublicKey: base64.RawURLEncoding.EncodeToString(fixture.registryPublic),
			NotBefore: fixtureNow.Add(-time.Hour).Format(time.RFC3339),
			SignUntil: fixtureNow.Add(time.Hour).Format(time.RFC3339)}, Retired: retired,
	}, fixture.releasePrivate)
	if err != nil {
		t.Fatal(err)
	}
	return bytes
}

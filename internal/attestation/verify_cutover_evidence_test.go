package attestation

import (
	"bytes"
	"testing"
	"time"
)

func TestVerifyCutoverEvidenceBuildsOpaqueProofChain(t *testing.T) {
	root, signer := cutoverTestAuthority(t)
	leaseDocument, signatureDocument, leaseSHA := signedCutoverLease(
		t, signer, attestationCutoverLease(),
	)
	lease, err := VerifyCutoverCredentialLease(
		leaseDocument, signatureDocument, root, cutoverTestNow,
	)
	if err != nil {
		t.Fatalf("verify lease: %v", err)
	}
	sealDocument := signCutoverSeal(t, signer, attestationSessionSeal(leaseSHA))
	seal, err := VerifyCutoverSessionSeal(sealDocument, lease, root, cutoverTestNow)
	if err != nil {
		t.Fatalf("verify session seal: %v", err)
	}
	cleanupDocument := signCutoverCleanup(
		t, signer, attestationCleanup(leaseSHA, seal.ReceiptSHA256()),
	)
	cleanup, err := VerifyCutoverCredentialCleanup(
		cleanupDocument, lease, seal, root, cutoverTestNow,
	)
	if err != nil {
		t.Fatalf("verify cleanup: %v", err)
	}
	if lease.LeaseID() != attestationLeaseID || lease.ManifestSHA256() != leaseSHA ||
		seal.ReceiptID() == "" || cleanup.ReceiptID() == "" || cleanup.ReceiptSHA256() == "" {
		t.Fatal("opaque cutover evidence lost its verified identity")
	}
}

func TestVerifyCutoverCredentialLeaseRejectsTamperAndExpiry(t *testing.T) {
	root, signer := cutoverTestAuthority(t)
	manifest, signature, _ := signedCutoverLease(t, signer, attestationCutoverLease())
	tampered := bytes.Replace(
		manifest,
		[]byte(`"plan_sha256":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"`),
		[]byte(`"plan_sha256":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"`),
		1,
	)
	if _, err := VerifyCutoverCredentialLease(tampered, signature, root, cutoverTestNow); err == nil {
		t.Fatal("tampered credential lease accepted")
	}
	if _, err := VerifyCutoverCredentialLease(
		manifest, signature, root, time.Date(2026, 7, 13, 12, 15, 0, 0, time.UTC),
	); err == nil {
		t.Fatal("expired credential lease accepted")
	}
	unknown := bytes.Replace(manifest, []byte(`{"contract_version"`), []byte(`{"unknown":true,"contract_version"`), 1)
	if _, err := VerifyCutoverCredentialLease(unknown, signature, root, cutoverTestNow); err == nil {
		t.Fatal("credential lease with unknown member accepted")
	}
}

func TestVerifyCutoverSessionSealRejectsUnboundOrTamperedReceipt(t *testing.T) {
	root, signer, lease, leaseSHA := verifiedCutoverLeaseFixture(t)
	if _, err := VerifyCutoverSessionSeal(
		signCutoverSeal(t, signer, attestationSessionSeal(leaseSHA)),
		VerifiedCutoverCredentialLease{}, root, cutoverTestNow,
	); err == nil {
		t.Fatal("session seal accepted without verified lease")
	}
	invalid := attestationSessionSeal(leaseSHA)
	invalid.Payload.SourceSessions[0].ApplicationName = "harpy-cutover/invalid/invalid"
	if _, err := VerifyCutoverSessionSeal(
		signCutoverSeal(t, signer, invalid), lease, root, cutoverTestNow,
	); err == nil {
		t.Fatal("session seal with unbound application name accepted")
	}
	document := signCutoverSeal(t, signer, attestationSessionSeal(leaseSHA))
	tampered := bytes.Replace(
		document,
		[]byte(`77777777-7777-4777-8777-777777777777`),
		[]byte(`99999999-9999-4999-8999-999999999999`), 1,
	)
	if _, err := VerifyCutoverSessionSeal(tampered, lease, root, cutoverTestNow); err == nil {
		t.Fatal("session seal with stale signature accepted")
	}
}

func TestVerifyCutoverCleanupRejectsNonTerminalEvidence(t *testing.T) {
	root, signer, lease, leaseSHA := verifiedCutoverLeaseFixture(t)
	sealDocument := signCutoverSeal(t, signer, attestationSessionSeal(leaseSHA))
	seal, err := VerifyCutoverSessionSeal(sealDocument, lease, root, cutoverTestNow)
	if err != nil {
		t.Fatal(err)
	}
	invalid := attestationCleanup(leaseSHA, seal.ReceiptSHA256())
	invalid.Payload.Principals[0].ActiveSessions = 1
	if _, err := VerifyCutoverCredentialCleanup(
		signCutoverCleanup(t, signer, invalid), lease, seal, root, cutoverTestNow,
	); err == nil {
		t.Fatal("cleanup receipt with active session accepted")
	}
	if _, err := VerifyCutoverCredentialCleanup(
		signCutoverCleanup(t, signer, attestationCleanup(leaseSHA, seal.ReceiptSHA256())),
		lease, VerifiedCutoverSessionSeal{}, root, cutoverTestNow,
	); err == nil {
		t.Fatal("cleanup receipt accepted without verified seal")
	}
	validDocument := signCutoverCleanup(
		t, signer, attestationCleanup(leaseSHA, seal.ReceiptSHA256()),
	)
	tampered := bytes.Replace(
		validDocument,
		[]byte(`88888888-8888-4888-8888-888888888888`),
		[]byte(`99999999-9999-4999-8999-999999999999`), 1,
	)
	if _, err := VerifyCutoverCredentialCleanup(
		tampered, lease, seal, root, cutoverTestNow,
	); err == nil {
		t.Fatal("cleanup receipt with stale signature accepted")
	}
}

func verifiedCutoverLeaseFixture(
	t *testing.T,
) (TrustRoot, Signer, VerifiedCutoverCredentialLease, string) {
	t.Helper()
	root, signer := cutoverTestAuthority(t)
	manifest, signature, digest := signedCutoverLease(t, signer, attestationCutoverLease())
	lease, err := VerifyCutoverCredentialLease(manifest, signature, root, cutoverTestNow)
	if err != nil {
		t.Fatal(err)
	}
	return root, signer, lease, digest
}

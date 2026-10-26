package attestation

import (
	"crypto/ed25519"
	"encoding/base64"
	"testing"
	"time"

	cutoverv1 "github.com/Onurlulardan/harpy/contracts/cutover/v1"
)

var cutoverTestNow = time.Date(2026, 7, 13, 12, 5, 0, 0, time.UTC)

func cutoverTestAuthority(t *testing.T) (TrustRoot, Signer) {
	t.Helper()
	seed := make([]byte, ed25519.SeedSize)
	for index := range seed {
		seed[index] = byte(index + 21)
	}
	privateKey := ed25519.NewKeyFromSeed(seed)
	rootDocument := mustReleaseJSON(t, map[string]any{
		"contract_version": "attestation-trust-root.v1",
		"keys": []any{map[string]any{
			"key_id": "cutover-1", "algorithm": "Ed25519",
			"public_key": base64.RawURLEncoding.EncodeToString(privateKey.Public().(ed25519.PublicKey)),
			"not_before": "2026-07-13T11:00:00Z", "not_after": "2026-07-13T14:00:00Z",
		}},
	})
	root, err := ParseTrustRoot(rootDocument)
	if err != nil {
		t.Fatalf("parse cutover trust root: %v", err)
	}
	signerDocument := mustReleaseJSON(t, map[string]any{
		"contract_version": "attestation-signing-key.v1", "key_id": "cutover-1",
		"algorithm": "Ed25519", "seed": base64.RawURLEncoding.EncodeToString(seed),
	})
	signer, err := ParseSigner(signerDocument, root)
	if err != nil {
		t.Fatalf("parse cutover signer: %v", err)
	}
	return root, signer
}

func signedCutoverLease(
	t *testing.T,
	signer Signer,
	lease cutoverv1.CredentialLease,
) ([]byte, []byte, string) {
	t.Helper()
	document := mustReleaseJSON(t, lease)
	issuedAt, _ := time.Parse(time.RFC3339Nano, lease.IssuedAt)
	signature, err := signer.SignAt(document, issuedAt)
	if err != nil {
		t.Fatal(err)
	}
	digest, err := DigestHex(document)
	if err != nil {
		t.Fatal(err)
	}
	signed := cutoverv1.CredentialLeaseSignature{
		ContractVersion: "cutover-credential-lease-signature.v1", ManifestSHA256: digest,
		Signature: cutoverv1.Signature{
			Algorithm: signature.Algorithm, KeyID: signature.KeyID, Value: signature.Value,
		},
	}
	return document, mustReleaseJSON(t, signed), digest
}

func signCutoverSeal(
	t *testing.T,
	signer Signer,
	receipt cutoverv1.SessionSealReceipt,
) []byte {
	t.Helper()
	issuedAt, _ := time.Parse(time.RFC3339Nano, receipt.Payload.IssuedAt)
	signature, err := signer.SignAt(mustReleaseJSON(t, receipt.SigningEnvelope()), issuedAt)
	if err != nil {
		t.Fatal(err)
	}
	receipt.Signature = cutoverv1.Signature{
		Algorithm: signature.Algorithm, KeyID: signature.KeyID, Value: signature.Value,
	}
	return mustReleaseJSON(t, receipt)
}

func signCutoverCleanup(
	t *testing.T,
	signer Signer,
	receipt cutoverv1.CredentialCleanupReceipt,
) []byte {
	t.Helper()
	issuedAt, _ := time.Parse(time.RFC3339Nano, receipt.Payload.IssuedAt)
	signature, err := signer.SignAt(mustReleaseJSON(t, receipt.SigningEnvelope()), issuedAt)
	if err != nil {
		t.Fatal(err)
	}
	receipt.Signature = cutoverv1.Signature{
		Algorithm: signature.Algorithm, KeyID: signature.KeyID, Value: signature.Value,
	}
	return mustReleaseJSON(t, receipt)
}

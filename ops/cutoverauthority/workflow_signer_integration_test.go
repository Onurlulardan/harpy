//go:build integration

package main

import (
	"crypto/ed25519"
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"
	"time"

	cutoverv1 "github.com/Onurlulardan/harpy/contracts/cutover/v1"
	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

type workflowSigner struct {
	root       attestation.TrustRoot
	signer     attestation.Signer
	rootPath   string
	signerPath string
}

func prepareWorkflowSigner(t *testing.T, directory string, now time.Time) workflowSigner {
	t.Helper()
	seed := make([]byte, ed25519.SeedSize)
	for index := range seed {
		seed[index] = 0x5a
	}
	privateKey := ed25519.NewKeyFromSeed(seed)
	keyID := "authority-integration"
	rootDocument := map[string]any{
		"contract_version": "attestation-trust-root.v1",
		"keys": []any{map[string]any{
			"key_id": keyID, "algorithm": "Ed25519",
			"public_key": base64.RawURLEncoding.EncodeToString(privateKey.Public().(ed25519.PublicKey)),
			"not_before": now.Add(-time.Hour).Format(time.RFC3339),
			"not_after":  now.Add(time.Hour).Format(time.RFC3339),
		}},
	}
	rootPath, rootBytes := writeWorkflowArtifact(t, directory, "trust-root.json", rootDocument)
	root, err := attestation.ParseTrustRoot(rootBytes)
	if err != nil {
		t.Fatal(err)
	}
	signerDocument := map[string]any{
		"contract_version": "attestation-signing-key.v1", "key_id": keyID,
		"algorithm": "Ed25519", "seed": base64.RawURLEncoding.EncodeToString(seed),
	}
	signerPath, signerBytes := writeWorkflowArtifact(
		t, directory, "signing-key.json", signerDocument,
	)
	signer, err := attestation.ParseSigner(signerBytes, root)
	if err != nil {
		t.Fatal(err)
	}
	return workflowSigner{root: root, signer: signer, rootPath: rootPath,
		signerPath: signerPath}
}

func signWorkflowDocument(
	t *testing.T, signer workflowSigner, document []byte, issuedAt time.Time,
	contractVersion string,
) []byte {
	t.Helper()
	digest, err := attestation.DigestHex(document)
	if err != nil {
		t.Fatal(err)
	}
	signed, err := signer.signer.SignAt(document, issuedAt)
	if err != nil {
		t.Fatal(err)
	}
	value := map[string]any{
		"contract_version": contractVersion, "manifest_sha256": digest,
		"signature": cutoverv1.Signature{
			Algorithm: signed.Algorithm, KeyID: signed.KeyID, Value: signed.Value,
		},
	}
	encoded, err := attestation.MarshalCanonical(value)
	if err != nil {
		t.Fatal(err)
	}
	return encoded
}

func runtimeWorkflowSignature(signature attestation.Signature) runtimev1.AttestationSignature {
	return runtimev1.AttestationSignature{
		Algorithm: signature.Algorithm, KeyID: signature.KeyID, Value: signature.Value,
	}
}

func writeWorkflowArtifact(
	t *testing.T, directory, name string, value any,
) (string, []byte) {
	t.Helper()
	document, err := attestation.MarshalCanonical(value)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(directory, name)
	if err := os.WriteFile(path, document, 0o600); err != nil {
		t.Fatal(err)
	}
	return path, document
}

func writeWorkflowBytes(t *testing.T, directory, name string, document []byte) string {
	t.Helper()
	path := filepath.Join(directory, name)
	if err := os.WriteFile(path, document, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

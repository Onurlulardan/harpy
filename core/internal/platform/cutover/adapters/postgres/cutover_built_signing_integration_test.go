//go:build integration

package postgres

import (
	"crypto/ed25519"
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"
	"time"

	cutoverv1 "github.com/Onurlulardan/harpy/contracts/cutover/v1"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

type builtCutoverArtifact struct {
	path     string
	document []byte
	digest   string
}

type builtCutoverSigner struct {
	signer         attestation.Signer
	root           attestation.TrustRoot
	rootPath       string
	signingKeyPath string
}

func prepareBuiltCutoverSigner(
	t *testing.T, directory string, now time.Time,
) builtCutoverSigner {
	t.Helper()
	seed := make([]byte, ed25519.SeedSize)
	for index := range seed {
		seed[index] = byte(91 + index)
	}
	privateKey := ed25519.NewKeyFromSeed(seed)
	root := writeBuiltCutoverArtifact(t, directory, "cutover-trust-root.json", map[string]any{
		"contract_version": "attestation-trust-root.v1",
		"keys": []any{map[string]any{
			"key_id": "cutover-built-gate", "algorithm": "Ed25519",
			"public_key": base64.RawURLEncoding.EncodeToString(
				privateKey.Public().(ed25519.PublicKey),
			),
			"not_before": now.Add(-time.Hour).Format(time.RFC3339),
			"not_after":  now.Add(time.Hour).Format(time.RFC3339),
		}},
	})
	parsedRoot, err := attestation.ParseTrustRoot(root.document)
	if err != nil {
		t.Fatalf("parse built cutover trust root: %v", err)
	}
	key, err := attestation.MarshalCanonical(map[string]any{
		"contract_version": "attestation-signing-key.v1",
		"key_id":           "cutover-built-gate", "algorithm": "Ed25519",
		"seed": base64.RawURLEncoding.EncodeToString(seed),
	})
	if err != nil {
		t.Fatal(err)
	}
	keyPath := filepath.Join(directory, "cutover-signing-key.json")
	if err := os.WriteFile(keyPath, key, 0o600); err != nil {
		t.Fatalf("write built cutover signing key: %v", err)
	}
	signer, err := attestation.ParseSigner(key, parsedRoot)
	if err != nil {
		t.Fatalf("parse built cutover signer: %v", err)
	}
	return builtCutoverSigner{
		signer: signer, root: parsedRoot, rootPath: root.path,
		signingKeyPath: keyPath,
	}
}

func writeBuiltCutoverArtifact(
	t *testing.T, directory string, name string, value any,
) builtCutoverArtifact {
	t.Helper()
	if validatable, ok := value.(interface{ Validate() error }); ok {
		if err := validatable.Validate(); err != nil {
			t.Fatalf("validate %s: %v", name, err)
		}
	}
	document, err := attestation.MarshalCanonical(value)
	if err != nil {
		t.Fatalf("canonicalize %s: %v", name, err)
	}
	path := filepath.Join(directory, name)
	if err := os.WriteFile(path, document, 0o600); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
	digest, err := attestation.DigestHex(document)
	if err != nil {
		t.Fatal(err)
	}
	return builtCutoverArtifact{path: path, document: document, digest: digest}
}

func publishBuiltCutoverArtifact(
	t *testing.T, path string, value any,
) builtCutoverArtifact {
	t.Helper()
	document, err := attestation.MarshalCanonical(value)
	if err != nil {
		t.Fatalf("canonicalize published cutover artifact: %v", err)
	}
	if err := attestation.WriteNewCanonical(path, document); err != nil {
		t.Fatalf("publish cutover artifact: %v", err)
	}
	digest, err := attestation.DigestHex(document)
	if err != nil {
		t.Fatal(err)
	}
	return builtCutoverArtifact{path: path, document: document, digest: digest}
}

func signBuiltCutoverDocument(
	t *testing.T, signer builtCutoverSigner, document []byte, issuedAt time.Time,
) cutoverv1.Signature {
	t.Helper()
	signature, err := signer.signer.SignAt(document, issuedAt)
	if err != nil {
		t.Fatalf("sign built cutover document: %v", err)
	}
	return cutoverv1.Signature{
		Algorithm: signature.Algorithm, KeyID: signature.KeyID, Value: signature.Value,
	}
}

func signBuiltCutoverEnvelope(
	t *testing.T, signer builtCutoverSigner, envelope any, issuedAt time.Time,
) cutoverv1.Signature {
	t.Helper()
	document, err := attestation.MarshalCanonical(envelope)
	if err != nil {
		t.Fatal(err)
	}
	return signBuiltCutoverDocument(t, signer, document, issuedAt)
}

//go:build integration

package migrator

import (
	"crypto/ed25519"
	"encoding/base64"
	"testing"
	"time"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

type builtEntrypointSigner struct {
	signer attestation.Signer
	root   builtEntrypointArtifact
}

func prepareBuiltEntrypointSigner(
	t *testing.T,
	directory string,
	now time.Time,
) builtEntrypointSigner {
	t.Helper()
	seed := make([]byte, ed25519.SeedSize)
	for index := range seed {
		seed[index] = byte(index + 41)
	}
	privateKey := ed25519.NewKeyFromSeed(seed)
	notBefore := now.Add(-time.Hour).Format(time.RFC3339)
	notAfter := now.Add(time.Hour).Format(time.RFC3339)
	root := writeBuiltEntrypointArtifact(t, directory, "trust-root.json", map[string]any{
		"contract_version": "attestation-trust-root.v1",
		"keys": []any{map[string]any{
			"key_id": "entrypoint-gate", "algorithm": "Ed25519",
			"public_key": base64.RawURLEncoding.EncodeToString(
				privateKey.Public().(ed25519.PublicKey),
			),
			"not_before": notBefore, "not_after": notAfter,
		}},
	})
	parsedRoot, err := attestation.ParseTrustRoot(root.document)
	if err != nil {
		t.Fatalf("parse built entrypoint trust root: %v", err)
	}
	signingKey, err := attestation.MarshalCanonical(map[string]any{
		"contract_version": "attestation-signing-key.v1",
		"key_id":           "entrypoint-gate", "algorithm": "Ed25519",
		"seed": base64.RawURLEncoding.EncodeToString(seed),
	})
	if err != nil {
		t.Fatalf("encode built entrypoint signing key: %v", err)
	}
	signer, err := attestation.ParseSigner(signingKey, parsedRoot)
	if err != nil {
		t.Fatalf("parse built entrypoint signer: %v", err)
	}
	return builtEntrypointSigner{signer: signer, root: root}
}

func builtEntrypointRuntimeSignature(
	t *testing.T,
	signer attestation.Signer,
	document []byte,
	issuedAt time.Time,
) map[string]any {
	t.Helper()
	signature, err := signer.SignAt(document, issuedAt)
	if err != nil {
		t.Fatalf("sign built entrypoint document: %v", err)
	}
	return map[string]any{
		"algorithm": signature.Algorithm,
		"key_id":    signature.KeyID,
		"value":     signature.Value,
	}
}

func builtEntrypointAttestationSignature(
	t *testing.T,
	signer attestation.Signer,
	document []byte,
	issuedAt time.Time,
) runtimev1.AttestationSignature {
	t.Helper()
	signature, err := signer.SignAt(document, issuedAt)
	if err != nil {
		t.Fatalf("sign built entrypoint receipt: %v", err)
	}
	return runtimev1.AttestationSignature{
		Algorithm: signature.Algorithm, KeyID: signature.KeyID, Value: signature.Value,
	}
}

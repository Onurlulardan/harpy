package worker

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSelectCustomerReleaseRootVerifiesBundleSignature(t *testing.T) {
	now := time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)
	releasePublic, releasePrivate, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	registryPublic, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	bundle := signedCustomerTrustBundle(t, now, registryPublic, releasePrivate)
	rootPath := writeCustomerTrustRoot(t, now, releasePublic)
	selected, err := selectCustomerReleaseRoot(rootPath, bundle, now)
	if err != nil || !selected.Equal(releasePublic) {
		t.Fatalf("valid release root selection failed: %v", err)
	}
	wrongPublic, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	wrongPath := writeCustomerTrustRoot(t, now, wrongPublic)
	if _, err := selectCustomerReleaseRoot(wrongPath, bundle, now); err == nil {
		t.Fatal("wrong release root accepted a signed SaaS trust bundle")
	}
}

func signedCustomerTrustBundle(
	t *testing.T,
	now time.Time,
	registryPublic ed25519.PublicKey,
	releasePrivate ed25519.PrivateKey,
) []byte {
	t.Helper()
	active := map[string]any{
		"algorithm": "Ed25519", "kid": "registry-1",
		"not_before": now.Add(-time.Hour).Format(time.RFC3339Nano),
		"public_key": base64.RawURLEncoding.EncodeToString(registryPublic),
		"sign_until": now.Add(time.Hour).Format(time.RFC3339Nano),
	}
	fields := map[string]any{
		"active": active, "registry_kind": "saas_provider", "retired": []any{},
		"schema": "protected-authority-ack-trust-bundle.v1", "trust_generation": int64(1),
	}
	body, err := json.Marshal(fields)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(body)
	signatureInput := append([]byte("harpy.protected-authority.ack-trust-bundle.v1\x00"), body...)
	fields["bundle_sha256"] = hex.EncodeToString(digest[:])
	fields["release_signature"] = base64.RawURLEncoding.EncodeToString(
		ed25519.Sign(releasePrivate, signatureInput),
	)
	document, err := json.Marshal(fields)
	if err != nil {
		t.Fatal(err)
	}
	return document
}

func writeCustomerTrustRoot(
	t *testing.T,
	now time.Time,
	publicKey ed25519.PublicKey,
) string {
	t.Helper()
	document := struct {
		ContractVersion string `json:"contract_version"`
		Keys            []struct {
			KeyID     string `json:"key_id"`
			Algorithm string `json:"algorithm"`
			PublicKey string `json:"public_key"`
			NotBefore string `json:"not_before"`
			NotAfter  string `json:"not_after"`
		} `json:"keys"`
	}{ContractVersion: "attestation-trust-root.v1"}
	document.Keys = append(document.Keys, struct {
		KeyID     string `json:"key_id"`
		Algorithm string `json:"algorithm"`
		PublicKey string `json:"public_key"`
		NotBefore string `json:"not_before"`
		NotAfter  string `json:"not_after"`
	}{
		KeyID: "release-1", Algorithm: "Ed25519",
		PublicKey: base64.RawURLEncoding.EncodeToString(publicKey),
		NotBefore: now.Add(-time.Hour).Format(time.RFC3339Nano),
		NotAfter:  now.Add(time.Hour).Format(time.RFC3339Nano),
	})
	wire, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "release-trust.json")
	if err := os.WriteFile(path, wire, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

package main

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadCatalogPackageAuthorityRequiresMatchingKeyring(t *testing.T) {
	seed := bytes.Repeat([]byte{0x51}, ed25519.SeedSize)
	privateKey := ed25519.NewKeyFromSeed(seed)
	publicKey := privateKey.Public().(ed25519.PublicKey)
	directory := t.TempDir()
	seedPath := filepath.Join(directory, "catalog-signing-seed")
	keyringPath := filepath.Join(directory, "catalog-keyring.json")
	if err := os.WriteFile(
		seedPath,
		[]byte(base64.RawURLEncoding.EncodeToString(seed)),
		0o600,
	); err != nil {
		t.Fatal(err)
	}
	writeCatalogKeyring(t, keyringPath, "catalog-release-2026-07", publicKey)

	authority, err := loadCatalogPackageAuthority(prepareOptions{
		catalogSigningSeed: seedPath,
		catalogKeyring:     keyringPath,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer clear(authority.privateKey)
	if authority.keyID != "catalog-release-2026-07" {
		t.Fatalf("unexpected signer key id %q", authority.keyID)
	}
	document := []byte("sealed Catalog package")
	signature, err := authority.Sign(document)
	if err != nil {
		t.Fatal(err)
	}
	resolved, err := authority.Resolve(authority.keyID)
	if err != nil {
		t.Fatal(err)
	}
	if !ed25519.Verify(resolved, document, signature) {
		t.Fatal("loaded Catalog package authority produced an invalid signature")
	}

	foreign := ed25519.NewKeyFromSeed(
		bytes.Repeat([]byte{0x52}, ed25519.SeedSize),
	).Public().(ed25519.PublicKey)
	writeCatalogKeyring(t, keyringPath, "catalog-release-2026-07", foreign)
	if _, err := loadCatalogPackageAuthority(prepareOptions{
		catalogSigningSeed: seedPath,
		catalogKeyring:     keyringPath,
	}); err == nil {
		t.Fatal("Catalog signing seed absent from the keyring was accepted")
	}
}

func TestLoadCatalogPackageAuthorityRejectsNonCanonicalSeed(t *testing.T) {
	directory := t.TempDir()
	seedPath := filepath.Join(directory, "catalog-signing-seed")
	keyringPath := filepath.Join(directory, "catalog-keyring.json")
	seed := bytes.Repeat([]byte{0x61}, ed25519.SeedSize)
	privateKey := ed25519.NewKeyFromSeed(seed)
	writeCatalogKeyring(
		t, keyringPath, "catalog-release-2026-07",
		privateKey.Public().(ed25519.PublicKey),
	)
	if err := os.WriteFile(
		seedPath,
		[]byte(base64.URLEncoding.EncodeToString(seed)),
		0o600,
	); err != nil {
		t.Fatal(err)
	}
	if _, err := loadCatalogPackageAuthority(prepareOptions{
		catalogSigningSeed: seedPath,
		catalogKeyring:     keyringPath,
	}); err == nil {
		t.Fatal("padded non-canonical Catalog signing seed was accepted")
	}
}

func writeCatalogKeyring(
	t *testing.T, path, keyID string, publicKey ed25519.PublicKey,
) {
	t.Helper()
	document := fmt.Sprintf(
		`{"version":1,"keys":[{"key_id":%q,"public_key":%q}]}`,
		keyID, base64.RawURLEncoding.EncodeToString(publicKey),
	)
	if err := os.WriteFile(path, []byte(document), 0o600); err != nil {
		t.Fatal(err)
	}
}

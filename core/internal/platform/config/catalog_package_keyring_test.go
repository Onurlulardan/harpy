package config

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"strings"
	"testing"
)

func TestCatalogPackageKeyringLoadsAndCopiesKeys(t *testing.T) {
	loaded, err := loadAPI(validAPIEnvironment())
	if err != nil {
		t.Fatalf("load API: %v", err)
	}
	keyring, err := loaded.CatalogPackageKeyring()
	if err != nil {
		t.Fatalf("catalog keyring: %v", err)
	}
	first, err := keyring.Resolve("release-2026-07")
	if err != nil || len(first) != ed25519.PublicKeySize {
		t.Fatal("resolve key failed")
	}
	first[0] = 0
	second, err := keyring.Resolve("release-2026-07")
	if err != nil || second[0] != 7 {
		t.Fatal("resolved key aliases keyring storage")
	}
	inventory := keyring.PublicKeys()
	if len(inventory) != 1 || inventory[0][0] != 7 {
		t.Fatal("catalog public key inventory differs")
	}
	inventory[0][0] = 0
	third, err := keyring.Resolve("release-2026-07")
	if err != nil || third[0] != 7 {
		t.Fatal("catalog public key inventory aliases keyring storage")
	}
}

func TestCatalogPackageKeyringRejectsAmbiguousAndInvalidDocuments(t *testing.T) {
	key := base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{8}, ed25519.PublicKeySize))
	tests := []string{
		`{"version":1,"version":1,"keys":[]}`,
		`{"version":1,"keys":[]}`,
		`{"version":1,"keys":[{"key_id":"A","public_key":"` + key + `"}]}`,
		`{"version":1,"keys":[{"key_id":"release","public_key":"bad"}]}`,
	}
	for _, document := range tests {
		entries := withEnvironment(validAPIEnvironment(),
			catalogPackageKeyringEnvironmentKey+"="+document)
		if _, err := loadAPI(entries); err == nil ||
			!strings.Contains(err.Error(), catalogPackageKeyringEnvironmentKey) {
			t.Fatalf("invalid keyring accepted: %s", document)
		}
	}
}

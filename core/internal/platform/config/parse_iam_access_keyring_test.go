package config

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestIAMAccessKeyringAcceptsRotationChain(t *testing.T) {
	now := time.Date(2026, 7, 2, 0, 0, 0, 0, time.UTC)
	document := iamAccessDocument(
		now, 10,
		iamRetiredAccessDocument(8, "2026-06-01T00:00:00Z", "2026-08-02T00:00:00Z"),
		iamRetiredAccessDocument(9, "2026-07-01T00:00:00Z", "2026-08-02T00:00:00Z"),
	)
	keyring, ok := parseIAMAccessKeyring(Secret{value: document}, now)
	if !ok || keyring.Validate() != nil {
		t.Fatal("valid access-token keyring rejected")
	}
	kid, privateKey, err := keyring.SigningKey()
	if err != nil || kid != iamFixtureKID(10) || len(privateKey) != ed25519.PrivateKeySize {
		t.Fatalf("unexpected signing key: kid=%q size=%d err=%v", kid, len(privateKey), err)
	}
	publicKeys, err := keyring.PublicKeys()
	if err != nil || len(publicKeys) != 3 {
		t.Fatalf("unexpected public key set: %d %v", len(publicKeys), err)
	}
}

func TestIAMAccessKeyringRejectsLifecycleAndWireDrift(t *testing.T) {
	now := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
	valid := iamAccessDocument(now, 10)
	tests := map[string]string{
		"duplicate member": strings.Replace(valid, `"version":1`, `"version":1,"version":1`, 1),
		"unknown member":   strings.Replace(valid, `"retired_public":[]`, `"retired_public":[],"extra":true`, 1),
		"future active":    validIAMAccessKeyringJSON(now.Add(-31*24*time.Hour), 10),
		"wrong interval":   strings.Replace(valid, "2026-08-13T00:00:00Z", "2026-08-12T00:00:00Z", 1),
		"padded seed":      strings.Replace(valid, iamFixtureMaterial(10), iamFixtureMaterial(10)+"=", 1),
		"noncanonical time": strings.Replace(
			valid, "2026-07-14T00:00:00Z", "2026-07-14T00:00:00+00:00", 1,
		),
		"invalid utf8": string(append([]byte(valid[:len(valid)-1]), 0xff, '}')),
	}
	for name, document := range tests {
		t.Run(name, func(t *testing.T) {
			if _, ok := parseIAMAccessKeyring(Secret{value: document}, now); ok {
				t.Fatal("invalid access-token keyring accepted")
			}
		})
	}
}

func TestIAMAccessKeyringRequiresContiguousUniqueRetiredKeys(t *testing.T) {
	now := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
	tests := []string{
		iamAccessDocument(now, 10,
			iamRetiredAccessDocument(8, "2026-05-31T00:00:00Z", "2026-08-02T00:00:00Z"),
			iamRetiredAccessDocument(9, "2026-07-01T00:00:00Z", "2026-08-02T00:00:00Z")),
		iamAccessDocument(now, 10,
			iamRetiredAccessDocument(9, "2026-07-01T00:00:00Z", "2026-08-02T00:00:00Z"),
			iamRetiredAccessDocument(9, "2026-07-31T00:00:00Z", "2026-09-02T00:00:00Z")),
		iamAccessDocument(now, 10,
			iamRetiredAccessDocument(9, "2026-07-01T00:00:00Z", "2026-07-01T12:00:00Z")),
	}
	for _, document := range tests {
		if _, ok := parseIAMAccessKeyring(Secret{value: document}, now); ok {
			t.Fatal("invalid retired access key chain accepted")
		}
	}
}

func TestIAMAccessKeyringReturnsDefensiveCopies(t *testing.T) {
	now := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
	keyring, ok := parseIAMAccessKeyring(Secret{value: iamAccessDocument(now, 10)}, now)
	if !ok {
		t.Fatal("fixture rejected")
	}
	_, first, _ := keyring.SigningKey()
	first[0] ^= 0xff
	_, second, _ := keyring.SigningKey()
	if bytes.Equal(first, second) {
		t.Fatal("signing private key aliases keyring storage")
	}
	publicKeys, _ := keyring.PublicKeys()
	publicKeys[iamFixtureKID(10)][0] ^= 0xff
	resolved, err := keyring.AccessTokenKey(t.Context(), iamFixtureKID(10))
	if err != nil || bytes.Equal(publicKeys[iamFixtureKID(10)], resolved) {
		t.Fatal("public key result aliases keyring storage")
	}
}

func iamAccessDocument(now time.Time, marker byte, retired ...string) string {
	base := validIAMAccessKeyringJSON(now, marker)
	return strings.Replace(base, `"retired_public":[]`, `"retired_public":[`+strings.Join(retired, ",")+`]`, 1)
}

func iamRetiredAccessDocument(marker byte, signUntil, publishUntil string) string {
	return fmt.Sprintf(
		`{"kid":%q,"public_key":%q,"sign_until":%q,"publish_until":%q}`,
		iamFixtureKID(marker), iamPublicMaterial(marker), signUntil, publishUntil,
	)
}

func iamPublicMaterial(marker byte) string {
	seed := bytes.Repeat([]byte{marker}, ed25519.SeedSize)
	privateKey := ed25519.NewKeyFromSeed(seed)
	publicKey := privateKey.Public().(ed25519.PublicKey)
	encoded := base64.RawURLEncoding.EncodeToString(publicKey)
	clear(privateKey)
	clear(seed)
	return encoded
}

package config

import (
	"crypto/ed25519"
	"encoding/base64"
	"os"
	"strings"
	"testing"
)

func TestActorBridgeSeedIsFileOnlyAndPurposeSeparated(t *testing.T) {
	entries := validAPIEnvironment()
	if _, err := loadAPI(append(entries,
		"HARPY_IAM_V1_ACTOR_BRIDGE_SIGNING_SEED=plaintext",
	)); err == nil {
		t.Fatalf("direct actor bridge seed error = %v", err)
	}
	loaded, err := loadAPI(entries)
	if err != nil {
		t.Fatal(err)
	}
	iam, _ := loaded.IAM()
	bridge, _ := iam.ActorBridge()
	_, key, _ := bridge.SigningKey()
	defer clear(key)
	if len(key) != 64 {
		t.Fatal("actor bridge signing key was not loaded")
	}
	access, _ := iam.AccessTokenKeyring()
	var accessSeed [32]byte
	for _, material := range access.keyMaterials() {
		accessSeed = material
		break
	}
	seedPath := environmentValue(entries, iamActorBridgeSeedFileKey)
	if err := os.WriteFile(seedPath, []byte(
		base64.RawURLEncoding.EncodeToString(accessSeed[:]),
	), 0o600); err != nil {
		t.Fatal(err)
	}
	accessPrivate := ed25519.NewKeyFromSeed(accessSeed[:])
	publicPath := environmentValue(entries, iamActorBridgePublicKeysFileKey)
	publicDocument := `{"version":1,"keys":[{"key_id":"edge-2026-07","public_key":"` +
		base64.RawURLEncoding.EncodeToString(accessPrivate[ed25519.SeedSize:]) + `"}]}`
	if err := os.WriteFile(publicPath, []byte(publicDocument), 0o600); err != nil {
		t.Fatal(err)
	}
	clear(accessPrivate)
	if _, err := loadAPI(entries); err == nil || !strings.Contains(err.Error(), "purpose") {
		t.Fatalf("reused access-token seed error = %v", err)
	}
}

func TestActorBridgeRejectsDuplicateJSONMembers(t *testing.T) {
	for _, fileKey := range []string{
		iamActorBridgePublicKeysFileKey,
		iamActorBridgeTargetsFileKey,
	} {
		entries := validAPIEnvironment()
		path := environmentValue(entries, fileKey)
		document, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		ambiguous := strings.Replace(string(document), `"version":1`,
			`"version":1,"version":1`, 1)
		if err := os.WriteFile(path, []byte(ambiguous), 0o600); err != nil {
			t.Fatal(err)
		}
		if _, err := loadAPI(entries); err == nil || !strings.Contains(err.Error(), "ambiguous JSON") {
			t.Fatalf("duplicate JSON member in %s error = %v", fileKey, err)
		}
	}
}

func TestActorBridgeCloseWipesStartupSigningKey(t *testing.T) {
	loaded, err := loadAPI(validAPIEnvironment())
	if err != nil {
		t.Fatal(err)
	}
	iam, _ := loaded.IAM()
	bridge, _ := iam.ActorBridge()
	_, key, err := bridge.SigningKey()
	if err != nil {
		t.Fatal(err)
	}
	clear(key)
	bridge.Close()
	if _, _, err := bridge.SigningKey(); err == nil {
		t.Fatal("signing key remained readable after close")
	}
}

func TestActorBridgeRequiresExactTargetsAndMatchingPublicKey(t *testing.T) {
	entries := validAPIEnvironment()
	targetPath := environmentValue(entries, iamActorBridgeTargetsFileKey)
	if err := os.WriteFile(targetPath, []byte(`{"version":1,"targets":[]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := loadAPI(entries); err == nil || !strings.Contains(err.Error(), "exact V1 domain") {
		t.Fatalf("incomplete target registry error = %v", err)
	}

	entries = validAPIEnvironment()
	publicPath := environmentValue(entries, iamActorBridgePublicKeysFileKey)
	wrong := base64.RawURLEncoding.EncodeToString(make([]byte, 32))
	document := `{"version":1,"keys":[{"key_id":"edge-2026-07","public_key":"` + wrong + `"}]}`
	if err := os.WriteFile(publicPath, []byte(document), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := loadAPI(entries); err == nil || !strings.Contains(err.Error(), "active signing key") {
		t.Fatalf("mismatched public set error = %v", err)
	}
}

func environmentValue(entries []string, key string) string {
	for _, entry := range entries {
		if strings.HasPrefix(entry, key+"=") {
			return strings.TrimPrefix(entry, key+"=")
		}
	}
	return ""
}

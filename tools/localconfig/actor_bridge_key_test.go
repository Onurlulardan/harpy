package main

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"testing"
)

func TestGeneratedActorBridgeSeedMatchesPublishedKeyAndExactTargets(t *testing.T) {
	set, _ := generatedSetForTest(t)
	seed, err := base64.RawURLEncoding.DecodeString(
		string(set.files["actor-bridge-signing-seed"]),
	)
	if err != nil || len(seed) != ed25519.SeedSize {
		t.Fatalf("invalid actor seed: %v", err)
	}
	defer clear(seed)
	var publicDocument struct {
		Keys []struct {
			KeyID     string `json:"key_id"`
			PublicKey string `json:"public_key"`
		} `json:"keys"`
	}
	if err := json.Unmarshal(set.files["actor-bridge-public-keys.json"],
		&publicDocument); err != nil || len(publicDocument.Keys) != 1 {
		t.Fatalf("invalid actor public document: %v", err)
	}
	privateKey := ed25519.NewKeyFromSeed(seed)
	defer clear(privateKey)
	want := base64.RawURLEncoding.EncodeToString(privateKey[ed25519.SeedSize:])
	if publicDocument.Keys[0].PublicKey != want ||
		publicDocument.Keys[0].KeyID != set.actorBridgeKeyID {
		t.Fatal("actor signing seed and public keyring differ")
	}
	var targets struct {
		Targets []struct {
			Domain     string `json:"domain"`
			BaseURL    string `json:"base_url"`
			ServerName string `json:"server_name"`
		} `json:"targets"`
	}
	if err := json.Unmarshal(set.files["actor-bridge-targets.json"], &targets); err != nil {
		t.Fatal(err)
	}
	if len(targets.Targets) != len(actorBridgeDomains) {
		t.Fatal("actor target domain set is incomplete")
	}
	for index, domain := range actorBridgeDomains {
		target := targets.Targets[index]
		if target.Domain != domain || target.ServerName != "v1-"+domain ||
			target.BaseURL != "https://v1-"+domain+":8443" {
			t.Fatalf("actor target differs for %s", domain)
		}
	}
}

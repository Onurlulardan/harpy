package config

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
)

func validIAMActorBridgeEnvironment() []string {
	directory, err := os.MkdirTemp("", "harpy-actor-bridge-config-")
	if err != nil {
		panic(err)
	}
	seed := bytes.Repeat([]byte{61}, ed25519.SeedSize)
	privateKey := ed25519.NewKeyFromSeed(seed)
	keyID := "edge-2026-07"
	paths := map[string]string{
		"seed":    filepath.Join(directory, "signing-seed"),
		"keys":    filepath.Join(directory, "public-keys.json"),
		"targets": filepath.Join(directory, "targets.json"),
		"cert":    filepath.Join(directory, "client.crt"),
		"private": filepath.Join(directory, "client.key"),
		"ca":      filepath.Join(directory, "server-ca.crt"),
	}
	publicDocument, _ := json.Marshal(map[string]any{
		"version": 1,
		"keys": []map[string]string{{
			"key_id": keyID,
			"public_key": base64.RawURLEncoding.EncodeToString(
				privateKey[ed25519.SeedSize:],
			),
		}},
	})
	targets := make([]map[string]string, 0, len(iamActorBridgeDomains))
	for _, domain := range iamActorBridgeDomains {
		host := "v1-" + string(domain) + ".fixture.invalid"
		targets = append(targets, map[string]string{
			"domain": string(domain), "base_url": "https://" + host,
			"server_name": host,
		})
	}
	targetDocument, _ := json.Marshal(map[string]any{"version": 1, "targets": targets})
	artifacts := map[string][]byte{
		paths["seed"]: []byte(base64.RawURLEncoding.EncodeToString(seed)),
		paths["keys"]: publicDocument, paths["targets"]: targetDocument,
		paths["cert"]:    []byte("fixture client certificate"),
		paths["private"]: []byte("fixture client private key"),
		paths["ca"]:      []byte("fixture server CA"),
	}
	for path, content := range artifacts {
		if err := os.WriteFile(path, content, 0o600); err != nil {
			panic(err)
		}
	}
	clear(seed)
	clear(privateKey)
	return []string{
		iamActorBridgeKeyIDKey + "=" + keyID,
		iamActorBridgeSeedFileKey + "=" + paths["seed"],
		iamActorBridgePublicKeysFileKey + "=" + paths["keys"],
		iamActorBridgeTargetsFileKey + "=" + paths["targets"],
		iamActorBridgeClientCertKey + "=" + paths["cert"],
		iamActorBridgeClientKeyKey + "=" + paths["private"],
		iamActorBridgeServerCAKey + "=" + paths["ca"],
	}
}

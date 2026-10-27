package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfigAcceptsCompleteLocalContract(t *testing.T) {
	values := validConfigValues(t)
	config, err := loadConfig(mapLookup(values))
	if err != nil {
		t.Fatalf("loadConfig: %v", err)
	}
	defer clear(config.masterKey)
	if config.listenAddr != "127.0.0.1:8443" || len(config.masterKey) != aesKeyBytes {
		t.Fatalf("unexpected config: %+v", config)
	}
}

func TestLoadConfigRejectsMissingValueAndWrongMasterKeySize(t *testing.T) {
	values := validConfigValues(t)
	delete(values, certKey)
	if _, err := loadConfig(mapLookup(values)); err == nil {
		t.Fatal("missing certificate was accepted")
	}

	values = validConfigValues(t)
	if err := os.WriteFile(values[masterKey], bytes.Repeat([]byte{1}, 31), 0o600); err != nil {
		t.Fatalf("replace master key: %v", err)
	}
	if _, err := loadConfig(mapLookup(values)); err == nil {
		t.Fatal("short master key was accepted")
	}
}

func TestLoadConfigRequiresSubjectTokenKeyForHostProfile(t *testing.T) {
	values := validConfigValues(t)
	values[profileKey] = "host"
	if _, err := loadConfig(mapLookup(values)); err == nil {
		t.Fatal("host profile accepted without subject-token key")
	}
	keyPath := filepath.Join(t.TempDir(), "subject-token.key")
	if err := os.WriteFile(keyPath, bytes.Repeat([]byte{0x42}, 32), 0o600); err != nil {
		t.Fatal(err)
	}
	values[tokenKey] = keyPath
	values[workloadKey] = writeWorkloadRegistryFixture(t)
	config, err := loadConfig(mapLookup(values))
	if err != nil {
		t.Fatal(err)
	}
	defer clear(config.masterKey)
	defer clear(config.tokenKey)
	if config.profile != "host" || len(config.tokenKey) != 32 ||
		len(config.workloads) != 2 {
		t.Fatalf("unexpected host config: profile=%q key=%d workloads=%d",
			config.profile, len(config.tokenKey), len(config.workloads))
	}
}

func writeWorkloadRegistryFixture(t *testing.T) string {
	t.Helper()
	workloads, _ := fixtureWorkloadRegistry()
	keys := make([]map[string]string, 0, len(workloads))
	for _, keyID := range []string{"api", "worker"} {
		workload := workloads[keyID]
		keys = append(keys, map[string]string{
			"kid": keyID, "subject": workload.subject,
			"public_key": base64.RawURLEncoding.EncodeToString(workload.publicKey),
		})
	}
	document, err := json.Marshal(map[string]any{"version": 1, "keys": keys})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "workloads.json")
	if err := os.WriteFile(path, document, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func validConfigValues(t *testing.T) map[string]string {
	t.Helper()
	certFile, keyFile := writeTLSFixture(t)
	keyPath := filepath.Join(t.TempDir(), "master.key")
	if err := os.WriteFile(keyPath, bytes.Repeat([]byte{0x73}, aesKeyBytes), 0o600); err != nil {
		t.Fatalf("write master key: %v", err)
	}
	return map[string]string{
		listenKey: "127.0.0.1:8443",
		certKey:   certFile,
		tlsKeyKey: keyFile,
		masterKey: keyPath,
	}
}

func mapLookup(values map[string]string) environmentLookup {
	return func(key string) (string, bool) {
		value, ok := values[key]
		return value, ok
	}
}

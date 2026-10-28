package main

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func secretFingerprints(t *testing.T, root, profile string) map[string]string {
	t.Helper()
	base := filepath.Join(root, "deploy", profile, "secrets")
	out := map[string]string{}
	err := filepath.WalkDir(base, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil || entry.IsDir() || !entry.Type().IsRegular() {
			return walkErr
		}
		rel, relErr := filepath.Rel(base, path)
		if relErr != nil {
			return relErr
		}
		out[filepath.ToSlash(rel)] = fileFingerprint(t, path)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(out) == 0 {
		t.Fatal("setup wrote no secret files")
	}
	return out
}

func fileFingerprint(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func assertSameFingerprints(t *testing.T, left, right map[string]string) {
	t.Helper()
	if len(left) != len(right) {
		t.Fatalf("secret file count = %d, want %d", len(right), len(left))
	}
	for path, sum := range left {
		if right[path] != sum {
			t.Fatalf("fingerprint changed for %s", path)
		}
	}
}

func stableAfterRotation(path string) bool {
	switch path {
	case "runtime/password-common.bin",
		"runtime/password-compromised.bin",
		"runtime/password-blocklists.json",
		"runtime/actor-bridge-targets.json",
		"runtime/actor-bridge-client.crt",
		"runtime/actor-bridge-client.key",
		"runtime/actor-bridge-server-ca.crt",
		"runtime/invitation-template.json",
		"runtime/installation.json",
		"runtime/recovery-trust.json",
		"runtime/ack-trust.json",
		"runtime/license.json",
		"runtime/license-trust.json",
		"runtime/release-trust.json",
		"runtime/authority-head.jsonl",
		"release/authority-manifest.json",
		"release/bootstrap-operation.json",
		"release/cleanup-operation.json",
		"release/cutover-operation.json",
		"release/extension-manifest.json",
		"release/extension-operation.json",
		"release/migration-manifest.json",
		"release/ownership-manifest.json",
		"release/privileges-manifest.json",
		"release/roles-manifest.json":
		return true
	default:
		return false
	}
}

func assertRotationMatrix(t *testing.T, before, after map[string]string) {
	t.Helper()
	if len(before) != len(after) {
		t.Fatalf("rotated file count = %d, want %d", len(after), len(before))
	}
	var kept []string
	for path, sum := range before {
		changed := after[path] != sum
		if stableAfterRotation(path) {
			if changed {
				t.Fatalf("rotate-secrets changed stable material %s", path)
			}
			continue
		}
		if !changed {
			kept = append(kept, path)
		}
	}
	if len(kept) > 0 {
		sort.Strings(kept)
		t.Fatalf("rotate-secrets kept %s", strings.Join(kept, ", "))
	}
}

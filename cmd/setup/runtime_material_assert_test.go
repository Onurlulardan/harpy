package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func assertUniqueEnvKeys(t *testing.T, lines []string) {
	t.Helper()
	seen := make(map[string]struct{}, len(lines))
	for _, line := range lines {
		key, value, ok := strings.Cut(line, "=")
		if !ok || key == "" || value == "" {
			t.Fatalf("malformed line %q", line)
		}
		if _, duplicate := seen[key]; duplicate {
			t.Fatalf("duplicate %s", key)
		}
		seen[key] = struct{}{}
	}
}

func assertOwnerOnlyRuntimeFiles(t *testing.T, root, profile string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		return
	}
	dir := filepath.Join(root, "deploy", profile, "secrets", "runtime")
	err := filepath.WalkDir(dir, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil || !entry.Type().IsRegular() {
			return walkErr
		}
		info, statErr := os.Stat(path)
		if statErr != nil {
			return statErr
		}
		if info.Mode().Perm() != 0o600 {
			t.Fatalf("%s mode is %o", path, info.Mode().Perm())
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func requireEnvKey(t *testing.T, lines []string, key string) {
	t.Helper()
	prefix := key + "="
	for _, line := range lines {
		if strings.HasPrefix(line, prefix) && len(line) > len(prefix) {
			return
		}
	}
	t.Fatalf("missing %s", key)
}

func withoutEnvKeys(lines []string, keys ...string) []string {
	drop := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		drop[key] = struct{}{}
	}
	kept := make([]string, 0, len(lines))
	for _, line := range lines {
		key, _, _ := strings.Cut(line, "=")
		if _, found := drop[key]; found {
			continue
		}
		kept = append(kept, line)
	}
	return kept
}

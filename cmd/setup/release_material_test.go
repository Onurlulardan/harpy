package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestReleaseMaterialWritesReleaseFiles(t *testing.T) {
	root := t.TempDir()
	lines, err := releaseMaterial(root, "single-tenant")
	if err != nil {
		t.Fatalf("release material: %v", err)
	}
	directory := filepath.Join(root, "deploy", "single-tenant", "secrets", "release")
	prefix := directory + string(os.PathSeparator)
	for _, line := range lines {
		key, path, found := strings.Cut(line, "=")
		info, statErr := os.Stat(path)
		if !found || !filepath.IsAbs(path) || !strings.HasPrefix(path, prefix) ||
			statErr != nil || info.Size() == 0 || !info.Mode().IsRegular() {
			t.Fatalf("%s was not written: %v", key, statErr)
		}
		if runtime.GOOS != "windows" && info.Mode().Perm() != 0o600 {
			t.Fatalf("%s mode = %o", key, info.Mode().Perm())
		}
	}
}

func TestReleaseMaterialRejectsUnknownProfile(t *testing.T) {
	if _, err := releaseMaterial(t.TempDir(), "saas"); err == nil {
		t.Fatal("unknown profile was accepted")
	}
}

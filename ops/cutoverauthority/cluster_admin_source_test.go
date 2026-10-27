package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestClusterAdminProductionPathNeverQueriesPGAuthID(t *testing.T) {
	rootOnly := map[string]bool{
		"quiesce_migrator.go":          true,
		"validate_adoption_catalog.go": true,
	}
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") ||
			strings.HasSuffix(name, "_test.go") || rootOnly[name] {
			continue
		}
		document, err := os.ReadFile(filepath.Clean(name))
		if err != nil {
			t.Fatal(err)
		}
		if bytes.Contains(bytes.ToLower(document), []byte("pg_authid")) {
			t.Fatalf("non-root authority source queries pg_authid: %s", name)
		}
	}
}

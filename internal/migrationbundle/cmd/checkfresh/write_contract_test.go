package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
)

func copyFreshTree(t *testing.T, repo, tree string) {
	t.Helper()
	err := filepath.WalkDir(filepath.Join(repo, "core"), func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(repo, path)
		if err != nil {
			return err
		}
		if !copyFreshInput(filepath.ToSlash(rel)) {
			return nil
		}
		return copyFile(path, filepath.Join(tree, rel))
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"binary-schema-contract.json", "fresh-migration-manifest.json"} {
		source := filepath.Join(repo, "contracts", "runtime", "v1", name)
		if err = copyFile(source, filepath.Join(tree, "contracts", "runtime", "v1", name)); err != nil {
			t.Fatal(err)
		}
	}
}

func copyFreshInput(slash string) bool {
	if strings.HasSuffix(slash, "/migration-supersessions.json") {
		return true
	}
	return strings.Contains(slash, "/migrations/") && strings.HasSuffix(slash, ".sql")
}

func mutateFreshSQL(t *testing.T, tree string) {
	t.Helper()
	target := filepath.Join(
		tree, "core", "internal", "platform", "migrator", "migrations", "fresh",
		"20260713010000_create_reject_mutation_routine.fresh.sql",
	)
	body := readBytes(t, target)
	updated := append(append([]byte{}, body...), []byte("-- checkfresh digest probe\n")...)
	if err := os.WriteFile(target, updated, 0o644); err != nil {
		t.Fatal(err)
	}
}

func copyFile(source, target string) error {
	body, err := os.ReadFile(source)
	if err != nil {
		return err
	}
	if err = os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	return os.WriteFile(target, body, 0o644)
}

func contractValue(t *testing.T, document []byte) runtimev1.BinarySchemaContract {
	t.Helper()
	var value runtimev1.BinarySchemaContract
	if err := json.Unmarshal(document, &value); err != nil {
		t.Fatal(err)
	}
	return value
}

func readBytes(t *testing.T, path string) []byte {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return body
}

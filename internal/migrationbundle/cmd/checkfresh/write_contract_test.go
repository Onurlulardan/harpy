package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/internal/migrationbundle"
)

func TestWriteUpdatesContractHashAndFreshManifest(t *testing.T) {
	repo, err := moduleRoot("")
	if err != nil {
		t.Fatal(err)
	}
	tree := t.TempDir()
	copyFreshTree(t, repo, tree)
	mutateFreshSQL(t, tree)

	contractBefore := readBytes(t, contractPath(tree))
	manifestBefore := readBytes(t, manifestPath(tree))
	if err = checkFresh(tree, false); err == nil {
		t.Fatal("stale fresh digest was accepted")
	}
	if !bytes.Equal(readBytes(t, contractPath(tree)), contractBefore) ||
		!bytes.Equal(readBytes(t, manifestPath(tree)), manifestBefore) {
		t.Fatal("mismatch wrote contract or manifest files")
	}

	if err = checkFresh(tree, true); err != nil {
		t.Fatal(err)
	}
	body, err := migrationbundle.CanonicalFresh(tree)
	if err != nil {
		t.Fatal(err)
	}
	written := contractValue(t, readBytes(t, contractPath(tree)))
	original := contractValue(t, contractBefore)
	if written.MigrationManifestSHA256 == original.MigrationManifestSHA256 {
		t.Fatal("contract migration hash was not updated")
	}
	if written.MigrationManifestSHA256 != migrationbundle.Digest(body) {
		t.Fatal("contract hash does not match the fresh manifest")
	}
	original.MigrationManifestSHA256 = written.MigrationManifestSHA256
	if written != original {
		t.Fatal("write changed a contract field other than the migration hash")
	}
	if !bytes.Equal(readBytes(t, manifestPath(tree)), body) ||
		bytes.Equal(body, manifestBefore) {
		t.Fatal("fresh manifest was not replaced from Fresh SQL")
	}
	if err = checkFresh(tree, false); err != nil {
		t.Fatal(err)
	}
}

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

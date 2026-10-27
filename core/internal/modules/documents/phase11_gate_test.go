package documents

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"testing"
)

func TestPhase11ModuleGate(t *testing.T) {
	root := documentsRepositoryRoot(t)
	assertDocumentMigrationsPresent(t, root)
	assertDocumentFileBudget(t, root)
	assertDocumentFragments(t, root)
}

func documentsRepositoryRoot(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve documents gate source")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(filename), "..", "..", "..", ".."))
}

func assertDocumentMigrationsPresent(t *testing.T, root string) {
	t.Helper()
	base := filepath.Join(root, "core", "internal", "modules", "documents", "migrations")
	info, err := os.Stat(base)
	if err != nil || !info.IsDir() {
		t.Fatalf("documents migrations directory: %v", err)
	}
}

func assertDocumentFileBudget(t *testing.T, root string) {
	t.Helper()
	paths := []string{filepath.Join(root, "core", "internal", "modules", "documents")}
	for _, relative := range documentOwnedFiles() {
		paths = append(paths, filepath.Join(root, filepath.FromSlash(relative)))
	}
	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if !info.IsDir() {
			assertDocumentLines(t, path)
			continue
		}
		err = filepath.WalkDir(path, func(filename string, entry os.DirEntry, walkErr error) error {
			if walkErr == nil && !entry.IsDir() && documentSource(entry.Name()) {
				assertDocumentLines(t, filename)
			}
			return walkErr
		})
		if err != nil {
			t.Fatal(err)
		}
	}
}

func documentOwnedFiles() []string {
	return []string{
		"contracts/authority/documents.json",
		"contracts/http/core-api-paths/documents.json",
		"contracts/http/documents-v1-compat.paths.json",
		"contracts/http/documents-v2.paths.json",
		"contracts/http/documents-v2.schemas.json",
		"contracts/operations/documents_adapters_postgres.json",
		"contracts/operations/documents_application.json",
		"contracts/storage/fragments/documents/01_external_accesses.json",
		"contracts/storage/migration-streams/documents.json",
		"contracts/storage/owners/documents.json",
	}
}

func documentSource(name string) bool {
	extension := filepath.Ext(name)
	return extension == ".go" || extension == ".sql" || extension == ".json" || extension == ".md"
}

func assertDocumentLines(t *testing.T, filename string) {
	t.Helper()
	document, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	lines := bytes.Count(document, []byte{'\n'})
	if len(document) > 0 && document[len(document)-1] != '\n' {
		lines++
	}
	if lines > 200 {
		t.Fatalf("documents file exceeds 200 lines: %s=%d", filename, lines)
	}
}

func assertDocumentFragments(t *testing.T, root string) {
	t.Helper()
	assertDocumentSortedCollection(t,
		filepath.Join(root, "contracts", "authority", "documents.json"), "authorities", "id")
	for _, name := range []string{"documents_adapters_postgres.json", "documents_application.json"} {
		assertDocumentSortedCollection(t,
			filepath.Join(root, "contracts", "operations", name), "operations", "source")
	}
	pathFile := filepath.Join(root, "contracts", "http", "core-api-paths", "documents.json")
	document, err := os.ReadFile(pathFile)
	var paths map[string]json.RawMessage
	if err != nil || json.Unmarshal(document, &paths) != nil || len(paths) != 25 {
		t.Fatalf("decode documents core paths: count=%d err=%v", len(paths), err)
	}
	for _, rootPath := range []string{"/api/document-archives", "/api/document-uploads", "/api/documents", "/api/folders"} {
		if _, found := paths[rootPath]; !found {
			t.Fatalf("documents core path root missing: %s", rootPath)
		}
	}
}

func assertDocumentSortedCollection(t *testing.T, filename, collection, key string) {
	t.Helper()
	document, err := os.ReadFile(filename)
	var raw map[string]json.RawMessage
	if err != nil || json.Unmarshal(document, &raw) != nil {
		t.Fatalf("decode fragment %s: %v", filename, err)
	}
	var entries []map[string]any
	if err = json.Unmarshal(raw[collection], &entries); err != nil || len(entries) == 0 {
		t.Fatalf("decode %s collection in %s: %v", collection, filename, err)
	}
	values := make([]string, len(entries))
	for index := range entries {
		values[index], _ = entries[index][key].(string)
	}
	if !sort.StringsAreSorted(values) {
		t.Fatalf("fragment %s is not sorted by %s", filename, key)
	}
	for index := 1; index < len(values); index++ {
		if values[index] == "" || values[index] == values[index-1] {
			t.Fatalf("fragment %s has empty or duplicate %s", filename, key)
		}
	}
}

package catalog

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
)

func TestPhase9ModuleGate(t *testing.T) {
	repository := catalogRepositoryRoot(t)
	assertCatalogMigrationPairs(t, repository, "tenant", 44)
	assertCatalogMigrationPairs(t, repository, "platform", 9)
	assertCatalogFileBudget(t, repository)
	assertCatalogFragments(t, repository)
}

func catalogRepositoryRoot(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve Catalog gate source")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(filename), "..", "..", "..", ".."))
}

func assertCatalogMigrationPairs(t *testing.T, root, scope string, want int) {
	t.Helper()
	base := filepath.Join(root, "core", "internal", "modules", "catalog", "migrations", scope)
	fresh, err := os.ReadDir(filepath.Join(base, "fresh"))
	if err != nil || len(fresh) != want {
		t.Fatalf("%s fresh migrations=%d want=%d err=%v", scope, len(fresh), want, err)
	}
	adopted, err := os.ReadDir(filepath.Join(base, "v1_adopted"))
	if err != nil || len(adopted) != want {
		t.Fatalf("%s adopted migrations=%d want=%d err=%v", scope, len(adopted), want, err)
	}
	byVersion := make(map[string]string, len(adopted))
	for _, entry := range adopted {
		byVersion[strings.SplitN(entry.Name(), "_", 2)[0]] = entry.Name()
	}
	for _, entry := range fresh {
		version := strings.SplitN(entry.Name(), "_", 2)[0]
		other, found := byVersion[version]
		if !found {
			t.Fatalf("%s migration %s lacks adopted pair", scope, entry.Name())
		}
		left, leftErr := os.ReadFile(filepath.Join(base, "fresh", entry.Name()))
		right, rightErr := os.ReadFile(filepath.Join(base, "v1_adopted", other))
		if leftErr != nil || rightErr != nil || !bytes.Equal(left, right) {
			t.Fatalf("%s migration pair %s differs: %v/%v", scope, version, leftErr, rightErr)
		}
	}
}

func assertCatalogFileBudget(t *testing.T, root string) {
	t.Helper()
	paths := []string{
		filepath.Join(root, "core", "internal", "modules", "catalog"),
		filepath.Join(root, "contracts", "http", "catalog-v2.paths.json"),
		filepath.Join(root, "contracts", "http", "catalog-v2.schemas.json"),
		filepath.Join(root, "contracts", "http", "core-api-paths", "catalog.json"),
	}
	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if !info.IsDir() {
			assertCatalogLines(t, path)
			continue
		}
		err = filepath.WalkDir(path, func(filename string, entry os.DirEntry, walkErr error) error {
			if walkErr == nil && !entry.IsDir() && catalogSource(entry.Name()) {
				assertCatalogLines(t, filename)
			}
			return walkErr
		})
		if err != nil {
			t.Fatal(err)
		}
	}
}

func catalogSource(name string) bool {
	extension := filepath.Ext(name)
	return extension == ".go" || extension == ".sql" || extension == ".json" || extension == ".md"
}

func assertCatalogLines(t *testing.T, filename string) {
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
		t.Fatalf("Catalog file exceeds 200 lines: %s=%d", filename, lines)
	}
}

func assertCatalogFragments(t *testing.T, root string) {
	t.Helper()
	for _, name := range []string{
		"catalog_adoption.json", "catalog_lifecycle.json", "catalog_menu.json",
		"catalog_operation.json", "catalog_registry.json",
	} {
		assertCatalogSortedFragment(t, filepath.Join(root, "contracts", "authority", name),
			"authorities", "id")
	}
	operations, err := filepath.Glob(filepath.Join(root, "contracts", "operations", "catalog_*.json"))
	if err != nil || len(operations) != 13 {
		t.Fatalf("Catalog operation fragments=%d err=%v", len(operations), err)
	}
	for _, filename := range operations {
		assertCatalogSortedFragment(t, filename, "operations", "source")
	}
	for _, filename := range []string{"catalog-v2.paths.json", "catalog-v2.schemas.json"} {
		var document map[string]any
		raw, readErr := os.ReadFile(filepath.Join(root, "contracts", "http", filename))
		if readErr != nil || json.Unmarshal(raw, &document) != nil {
			t.Fatalf("decode Catalog OpenAPI fragment %s: %v", filename, readErr)
		}
	}
}

func assertCatalogSortedFragment(t *testing.T, filename, collection, key string) {
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

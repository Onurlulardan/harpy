package runtimev1

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWorkOpenAPIPathsArePublishedInTheCoreAggregate(t *testing.T) {
	directory := filepath.Join("..", "..", "http")
	expected := workFragmentReferences(t, directory)
	fragment := readOpenAPIDocument(t, filepath.Join(directory, "core-api-paths", "work.json"))
	if len(fragment) != len(expected) {
		t.Fatalf("work catalog paths = %d, want %d", len(fragment), len(expected))
	}
	core := readOpenAPIDocument(t, filepath.Join(directory, "core-api-v1.openapi.yaml"))
	paths := core["paths"].(map[string]any)
	for path, reference := range expected {
		catalog, found := fragment[path].(map[string]any)
		if !found || len(catalog) != 1 || catalog["$ref"] != reference {
			t.Fatalf("work catalog reference for %s = %v", path, fragment[path])
		}
		published, found := paths[path].(map[string]any)
		if !found || len(published) != 1 || published["$ref"] != reference {
			t.Fatalf("core aggregate reference for %s = %v", path, paths[path])
		}
		if !workReferenceResolves(t, directory, reference) {
			t.Fatalf("work reference does not resolve: %s", reference)
		}
	}
}

func workFragmentReferences(t *testing.T, directory string) map[string]string {
	t.Helper()
	references := map[string]string{}
	for _, name := range []string{"work-v2.paths.json", "work-v1-compat.paths.json"} {
		document := readOpenAPIDocument(t, filepath.Join(directory, name))
		operations, ok := document["paths"].(map[string]any)
		if !ok || len(operations) == 0 {
			t.Fatalf("%s has no paths", name)
		}
		for path := range operations {
			if _, duplicate := references[path]; duplicate {
				t.Fatalf("work path %s is owned twice", path)
			}
			references[path] = name + "#/paths/" + workJSONPointer(path)
		}
	}
	return references
}

func workReferenceResolves(t *testing.T, directory, reference string) bool {
	t.Helper()
	file, pointer, ok := strings.Cut(reference, "#")
	if !ok || !strings.HasPrefix(pointer, "/paths/") {
		return false
	}
	content, err := os.ReadFile(filepath.Join(directory, file))
	if err != nil {
		t.Fatal(err)
	}
	var document map[string]json.RawMessage
	if err = json.Unmarshal(content, &document); err != nil {
		t.Fatal(err)
	}
	var paths map[string]json.RawMessage
	if err = json.Unmarshal(document["paths"], &paths); err != nil {
		t.Fatal(err)
	}
	_, found := paths[workUnescapePointer(strings.TrimPrefix(pointer, "/paths/"))]
	return found
}

func workJSONPointer(path string) string {
	return strings.ReplaceAll(strings.ReplaceAll(path, "~", "~0"), "/", "~1")
}

func workUnescapePointer(token string) string {
	return strings.ReplaceAll(strings.ReplaceAll(token, "~1", "/"), "~0", "~")
}

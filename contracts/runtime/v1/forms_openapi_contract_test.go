package runtimev1

import (
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestFormsCanonicalOpenAPIRoutesAreExactAndSecured(t *testing.T) {
	directory := filepath.Join("..", "..", "http")
	core := readOpenAPIDocument(t, filepath.Join(directory, "core-api-v1.openapi.yaml"))
	paths := core["paths"].(map[string]any)
	wanted := map[string][]string{
		"/api/forms":                       {"get", "post"},
		"/api/forms/{form_id}":             {"get"},
		"/api/forms/{form_id}/draft":       {"patch"},
		"/api/forms/{form_id}/versions":    {"post"},
		"/api/forms/{form_id}/retirements": {"post"},
		"/api/forms/{form_id}/versions/{form_version_id}/publications":                                      {"post"},
		"/api/forms/{form_id}/versions/{form_version_id}/publications/{publication_id}":                     {"get"},
		"/api/forms/{form_id}/versions/{form_version_id}/publications/{publication_id}/validations":         {"post"},
		"/api/forms/{form_id}/versions/{form_version_id}/publications/{publication_id}/plans":               {"post"},
		"/api/forms/{form_id}/versions/{form_version_id}/publications/{publication_id}/activations":         {"post"},
		"/api/forms/{form_id}/versions/{form_version_id}/publications/{publication_id}/deactivations":       {"post"},
		"/api/forms/{form_id}/versions/{form_version_id}/publications/{publication_id}/records":             {"post"},
		"/api/forms/{form_id}/versions/{form_version_id}/publications/{publication_id}/records/query":       {"post"},
		"/api/forms/{form_id}/versions/{form_version_id}/publications/{publication_id}/records/{record_id}": {"delete", "get", "patch"},
		"/api/forms/{form_id}/record-exports":                                                               {"post"},
		"/api/forms/{form_id}/record-exports/{export_id}":                                                   {"delete", "get"},
		"/api/forms/{form_id}/record-exports/{export_id}/content":                                           {"get"},
		"/api/forms/{form_id}/record-import-sources/{source_id}":                                            {"put"},
		"/api/forms/{form_id}/record-import-sources/{source_id}/analysis":                                   {"get"},
		"/api/forms/{form_id}/record-imports":                                                               {"post"},
		"/api/forms/{form_id}/record-imports/{import_id}":                                                   {"delete", "get"},
		"/api/forms/{form_id}/record-imports/{import_id}/errors":                                            {"get"},
	}
	for path, methods := range wanted {
		item := resolveFormsPathItem(t, directory, paths[path])
		assertFormsTenantHeaders(t, path, item)
		var found []string
		for _, method := range []string{"delete", "get", "patch", "post", "put"} {
			operation, exists := item[method].(map[string]any)
			if !exists {
				continue
			}
			found = append(found, method)
			if _, secured := operation["security"]; !secured ||
				operation["operationId"] == "" || operation["x-harpy-permission"] == "" ||
				operation["x-harpy-authority"] == "" {
				t.Fatalf("Forms operation is incomplete: %s %s %#v", method, path, operation)
			}
			responses, responsesValid := operation["responses"].(map[string]any)
			if !responsesValid {
				t.Fatalf("Forms responses are missing: %s %s", method, path)
			}
			if _, exactErrors := responses["default"]; !exactErrors && len(responses) < 2 {
				t.Fatalf("Forms error response contract is missing: %s %s", method, path)
			}
		}
		sort.Strings(methods)
		if strings.Join(found, ",") != strings.Join(methods, ",") {
			t.Fatalf("Forms methods for %s = %v, want %v", path, found, methods)
		}
	}
}

func assertFormsTenantHeaders(t *testing.T, path string, item map[string]any) {
	t.Helper()
	parameters, valid := item["parameters"].([]any)
	if !valid {
		t.Fatalf("Forms path has no tenant header contract: %s", path)
	}
	foundID, foundSlug := false, false
	for _, value := range parameters {
		parameter, ok := value.(map[string]any)
		if !ok {
			continue
		}
		reference, _ := parameter["$ref"].(string)
		foundID = foundID || strings.HasSuffix(reference, "/TenantIDParameter")
		foundSlug = foundSlug || strings.HasSuffix(reference, "/TenantSlugParameter")
	}
	if !foundID || !foundSlug {
		t.Fatalf("Forms tenant headers are incomplete for %s", path)
	}
}

func resolveFormsPathItem(
	t *testing.T,
	directory string,
	value any,
) map[string]any {
	t.Helper()
	item, valid := value.(map[string]any)
	if !valid {
		t.Fatalf("Forms path item has invalid shape: %#v", value)
	}
	reference, referenced := item["$ref"].(string)
	if !referenced {
		return item
	}
	parts := strings.SplitN(reference, "#", 2)
	document := readOpenAPIDocument(t, filepath.Join(directory, filepath.FromSlash(parts[0])))
	if len(parts) != 2 || !strings.HasPrefix(parts[1], "/") {
		t.Fatalf("Forms path reference is invalid: %s", reference)
	}
	var current any = document
	for _, token := range strings.Split(strings.TrimPrefix(parts[1], "/"), "/") {
		name := strings.ReplaceAll(strings.ReplaceAll(token, "~1", "/"), "~0", "~")
		object, ok := current.(map[string]any)
		if !ok {
			t.Fatalf("Forms path reference cannot traverse %s", reference)
		}
		current, ok = object[name]
		if !ok {
			t.Fatalf("Forms path reference is unresolved: %s", reference)
		}
	}
	resolved, ok := current.(map[string]any)
	if !ok {
		t.Fatalf("Forms path reference is not an object: %s", reference)
	}
	return resolved
}

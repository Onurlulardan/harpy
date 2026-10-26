package runtimev1

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestFormsV1DefinitionOpenAPIHasExactClosedInventory(t *testing.T) {
	document := readFormsV1DefinitionContract(t,
		"forms-v1-definition-compat.paths.json")
	items := document["pathItems"].(map[string]any)
	want := map[string][]string{
		"collection": {"get", "post"}, "item": {"get", "put", "delete"},
		"bySlug": {"get"}, "validate": {"post"}, "checkSlug": {"get"},
		"versions": {"get"}, "publish": {"post"}, "activate": {"post"},
		"unpublish": {"post"},
	}
	operations := 0
	if len(items) != len(want) {
		t.Fatalf("path items=%d", len(items))
	}
	for name, methods := range want {
		item, found := items[name].(map[string]any)
		if !found || len(item) != len(methods)+1 {
			t.Fatalf("path item %s is invalid", name)
		}
		parameters, ok := item["parameters"].([]any)
		if !ok || len(parameters) != 2 ||
			parameterRef(parameters[0]) != "forms-v1-definition-compat.schemas.json#/parameters/TenantIDParameter" ||
			parameterRef(parameters[1]) != "forms-v1-definition-compat.schemas.json#/parameters/TenantSlugParameter" {
			t.Fatalf("path item %s tenant parameters are incomplete", name)
		}
		for _, method := range methods {
			operation, ok := item[method].(map[string]any)
			responses, responsesOK := operation["responses"].(map[string]any)
			if !ok || operation["deprecated"] != true || operation["operationId"] == "" ||
				operation["x-harpy-permission"] == nil ||
				operation["x-harpy-authority"] == nil || !responsesOK ||
				responses["default"] == nil {
				t.Fatalf("%s %s is not a deprecated operation", method, name)
			}
			operations++
		}
	}
	if operations != 12 {
		t.Fatalf("operation count=%d", operations)
	}
}

func TestFormsV1DefinitionOpenAPIDeclaresIntentionalDifferences(t *testing.T) {
	document := readFormsV1DefinitionContract(t,
		"forms-v1-definition-compat.paths.json")
	items := document["pathItems"].(map[string]any)
	for _, name := range []string{"collection", "item", "validate", "checkSlug",
		"publish", "activate", "unpublish"} {
		item := items[name].(map[string]any)
		found := false
		for key, value := range item {
			if key == "parameters" {
				continue
			}
			operation := value.(map[string]any)
			if operation["x-harpy-compat-policy"] != nil ||
				operation["x-harpy-compat-owner"] != nil {
				found = true
			}
		}
		if !found {
			t.Fatalf("path item %s has no compatibility policy", name)
		}
	}
}

func parameterRef(value any) string {
	parameter, ok := value.(map[string]any)
	if !ok {
		return ""
	}
	reference, _ := parameter["$ref"].(string)
	return reference
}

func readFormsV1DefinitionContract(t *testing.T, name string) map[string]any {
	t.Helper()
	path := filepath.Join("..", "..", "http", name)
	document, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err = json.Unmarshal(document, &decoded); err != nil {
		t.Fatal(err)
	}
	return decoded
}

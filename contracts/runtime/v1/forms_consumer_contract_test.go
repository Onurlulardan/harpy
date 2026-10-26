package runtimev1

import (
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
	"testing"

	formsv1 "github.com/Onurlulardan/harpy/contracts/consumers/formsv1"
)

func TestFormsV1ConsumerContractMatchesOpenAPI(t *testing.T) {
	contract, err := formsv1.Load()
	if err != nil {
		t.Fatal(err)
	}
	directory := filepath.Join("..", "..", "http")
	core := readOpenAPIDocument(t, filepath.Join(directory, "core-api-v1.openapi.yaml"))
	paths := core["paths"].(map[string]any)
	schemas := readOpenAPIDocument(t, filepath.Join(
		directory, "forms-v1-compat.schemas.json",
	))
	assertConsumerSuccessSchema(t, contract, schemas)
	for _, route := range contract.Routes {
		item := resolveFormsPathItem(t, directory, paths[route.Path])
		operation, valid := item[strings.ToLower(route.Method)].(map[string]any)
		if !valid || operation["operationId"] != route.OperationID ||
			operation["deprecated"] != true || len(operation["security"].([]any)) == 0 {
			t.Fatalf("consumer operation drifted: %s %s", route.Method, route.Path)
		}
		assertConsumerMethods(t, route, item)
		assertConsumerResponses(t, route, operation)
		assertConsumerQueryKeys(t, contract, route, operation, schemas)
	}
}

func assertConsumerSuccessSchema(
	t *testing.T, contract formsv1.Contract, schemas map[string]any,
) {
	t.Helper()
	definitions := schemas["schemas"].(map[string]any)
	root := definitions["ListSuccess"].(map[string]any)
	for _, pointer := range contract.SuccessPointers {
		current := root
		for _, token := range strings.Split(strings.TrimPrefix(pointer, "/"), "/") {
			current = resolveConsumerSchemaReference(t, current, definitions)
			required := current["required"].([]any)
			if !containsSchemaName(required, token) {
				t.Fatalf("consumer pointer is not OpenAPI-required: %s", pointer)
			}
			properties := current["properties"].(map[string]any)
			current = properties[token].(map[string]any)
		}
	}
}

func resolveConsumerSchemaReference(
	t *testing.T, schema map[string]any, definitions map[string]any,
) map[string]any {
	t.Helper()
	reference, found := schema["$ref"].(string)
	if !found {
		return schema
	}
	const prefix = "#/schemas/"
	if !strings.HasPrefix(reference, prefix) {
		t.Fatalf("consumer schema reference is external: %s", reference)
	}
	resolved, valid := definitions[strings.TrimPrefix(reference, prefix)].(map[string]any)
	if !valid {
		t.Fatalf("consumer schema reference is unresolved: %s", reference)
	}
	return resolved
}

func containsSchemaName(values []any, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}

func assertConsumerMethods(t *testing.T, route formsv1.Route, item map[string]any) {
	t.Helper()
	var actual []string
	for _, method := range []string{"delete", "get", "head", "patch", "post", "put"} {
		if _, found := item[method]; found {
			actual = append(actual, strings.ToUpper(method))
		}
	}
	if !slices.Equal(actual, []string{route.Method}) {
		t.Fatalf("consumer methods for %s = %v", route.Path, actual)
	}
}

func assertConsumerResponses(
	t *testing.T, route formsv1.Route, operation map[string]any,
) {
	t.Helper()
	responses, valid := operation["responses"].(map[string]any)
	if !valid {
		t.Fatalf("consumer responses missing: %s", route.Path)
	}
	expectedSet := make(map[string]bool, len(route.Errors)+1)
	if route.SuccessStatus != 0 {
		expectedSet[strconv.Itoa(route.SuccessStatus)] = true
	}
	for _, contract := range route.Errors {
		expectedSet[strconv.Itoa(contract.Status)] = true
	}
	actual := make([]string, 0, len(responses))
	expected := make([]string, 0, len(expectedSet))
	for status := range responses {
		actual = append(actual, status)
	}
	for status := range expectedSet {
		expected = append(expected, status)
	}
	sort.Strings(actual)
	sort.Strings(expected)
	if !slices.Equal(actual, expected) {
		t.Fatalf("consumer responses for %s = %v, want %v", route.Path, actual, expected)
	}
}

func assertConsumerQueryKeys(
	t *testing.T, contract formsv1.Contract, route formsv1.Route,
	operation map[string]any, schemas map[string]any,
) {
	t.Helper()
	parameters, valid := operation["parameters"].([]any)
	if !valid {
		t.Fatalf("consumer parameters missing: %s", route.Path)
	}
	definitions := schemas["parameters"].(map[string]any)
	var query []string
	for _, value := range parameters {
		parameter := value.(map[string]any)
		reference := parameter["$ref"].(string)
		name := reference[strings.LastIndex(reference, "/")+1:]
		resolved := definitions[name].(map[string]any)
		if resolved["in"] == "query" {
			query = append(query, resolved["name"].(string))
		}
	}
	sort.Strings(query)
	expected := []string{}
	if route.ID == "records.list" {
		expected = contract.QueryKeys
	}
	if !slices.Equal(query, expected) {
		t.Fatalf("consumer query keys for %s = %v, want %v", route.Path, query, expected)
	}
}

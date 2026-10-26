package cutoverv1

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

func TestRegistryMatchesExactCutoverOrder(t *testing.T) {
	schema := compileCutoverSchema(t, "registry.schema.json")
	registry := readCutoverJSON(t, "registry.json")
	if err := schema.Validate(registry); err != nil {
		t.Fatalf("cutover registry rejected: %v", err)
	}
}

func TestRegistryRejectsPartialBillingFormsGroup(t *testing.T) {
	registry := readCutoverJSON(t, "registry.json").(map[string]any)
	groups := registry["groups"].([]any)
	groups[2].(map[string]any)["members"] = []any{"billing"}
	if err := compileCutoverSchema(t, "registry.schema.json").Validate(registry); err == nil {
		t.Fatal("partial billing+forms group accepted")
	}
}

func TestRegistryRejectsPartialIAMSystemResourcesGroup(t *testing.T) {
	registry := readCutoverJSON(t, "registry.json").(map[string]any)
	groups := registry["groups"].([]any)
	groups[1].(map[string]any)["members"] = []any{"iam"}
	if err := compileCutoverSchema(t, "registry.schema.json").Validate(registry); err == nil {
		t.Fatal("partial IAM system resources group accepted")
	}
}

func TestRegistryRejectsLegacyIAMCommandGroup(t *testing.T) {
	registry := readCutoverJSON(t, "registry.json").(map[string]any)
	groups := registry["groups"].([]any)
	groups[1].(map[string]any)["group_id"] = "iam"
	if err := compileCutoverSchema(t, "registry.schema.json").Validate(registry); err == nil {
		t.Fatal("legacy IAM command group accepted")
	}
}

func compileCutoverSchema(t *testing.T, path string) *jsonschema.Schema {
	t.Helper()
	compiler := jsonschema.NewCompiler()
	entries, err := filepath.Glob("*.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	const base = "https://harpy.io/contracts/cutover/v1/"
	for _, entry := range entries {
		file, openErr := os.Open(entry)
		if openErr != nil {
			t.Fatal(openErr)
		}
		value, decodeErr := jsonschema.UnmarshalJSON(file)
		_ = file.Close()
		if decodeErr != nil {
			t.Fatalf("decode %s: %v", entry, decodeErr)
		}
		if err := compiler.AddResource(base+entry, value); err != nil {
			t.Fatal(err)
		}
	}
	schema, err := compiler.Compile(base + path)
	if err != nil {
		t.Fatal(err)
	}
	return schema
}

func readCutoverJSON(t *testing.T, path string) any {
	t.Helper()
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	value, err := jsonschema.UnmarshalJSON(file)
	if err != nil {
		t.Fatal(err)
	}
	return value
}

package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

const runtimeSchemaBase = "https://harpy.io/contracts/runtime/v1/"

func compileRuntimeSchema(t *testing.T, name string) *jsonschema.Schema {
	t.Helper()
	compiler := jsonschema.NewCompiler()
	directory := filepath.Join("..", "..", "contracts", "runtime", "v1")
	entries, err := filepath.Glob(filepath.Join(directory, "*.schema.json"))
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		file, err := os.Open(entry)
		if err != nil {
			t.Fatal(err)
		}
		var document any
		decodeErr := json.NewDecoder(file).Decode(&document)
		closeErr := file.Close()
		if decodeErr != nil || closeErr != nil {
			t.Fatalf("read schema %s: decode=%v close=%v", entry, decodeErr, closeErr)
		}
		resource := runtimeSchemaBase + filepath.Base(entry)
		if err := compiler.AddResource(resource, document); err != nil {
			t.Fatalf("register schema %s: %v", entry, err)
		}
	}
	schema, err := compiler.Compile(runtimeSchemaBase + name)
	if err != nil {
		t.Fatalf("compile schema %s: %v", name, err)
	}
	return schema
}

func decodeJSONArtifact(t *testing.T, content []byte) any {
	t.Helper()
	var value any
	if err := json.Unmarshal(content, &value); err != nil {
		t.Fatalf("decode generated JSON: %v", err)
	}
	return value
}

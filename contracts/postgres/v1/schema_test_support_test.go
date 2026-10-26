package postgresv1

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

const schemaBase = "https://harpy.io/contracts/postgres/v1/"

const postgresImageDigest = "sha256:c8a919765f2ef63681329fa21021b830cd4d79d1165bdca730dd016014e4da84"

var postgresSchemaNames = []string{
	"postgres-artifact-defs.schema.json",
	"postgres-operation-defs.schema.json",
	"migration-manifest.schema.json",
	"extension-manifest.schema.json",
	"roles-manifest.schema.json",
	"privileges-manifest.schema.json",
	"ownership-manifest.schema.json",
	"ownership-chunk.schema.json",
	"bootstrap-operation.schema.json",
	"extension-operation.schema.json",
	"credential-cleanup-operation.schema.json",
}

func compileArtifactSchema(t *testing.T, name string) *jsonschema.Schema {
	t.Helper()
	compiler := jsonschema.NewCompiler()
	files, err := filepath.Glob("*.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range files {
		value := readSchemaJSON(t, path)
		identifier, ok := value.(map[string]any)["$id"].(string)
		if !ok {
			t.Fatalf("schema %s has no $id", path)
		}
		if err := compiler.AddResource(identifier, value); err != nil {
			t.Fatalf("register schema %s: %v", path, err)
		}
	}
	schema, err := compiler.Compile(schemaBase + name)
	if err != nil {
		t.Fatalf("compile schema %s: %v", name, err)
	}
	return schema
}

func readSchemaJSON(t *testing.T, path string) any {
	t.Helper()
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	value, err := jsonschema.UnmarshalJSON(file)
	if err != nil {
		t.Fatalf("decode schema %s: %v", path, err)
	}
	return value
}

func cloneDocument(t *testing.T, value map[string]any) map[string]any {
	t.Helper()
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := jsonschema.UnmarshalJSON(bytes.NewReader(encoded))
	if err != nil {
		t.Fatal(err)
	}
	return decoded.(map[string]any)
}

func digest(char string) string { return strings.Repeat(char, 64) }

func operationStep(id, class string) map[string]any {
	return map[string]any{"step_id": id, "transaction_class": class}
}

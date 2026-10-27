package effectsv1

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

const schemaBase = "https://harpy.io/contracts/effects/v1/"

func TestEffectSchemasCompileAndValidateContracts(t *testing.T) {
	failed, err := NewFailedReceipt(
		requestFixture(t), testReceiptID, "RATE_LIMITED", ErrorTransient,
		nil, "2026-07-15T10:00:01Z",
	)
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name   string
		schema string
		value  any
	}{
		{name: "request", schema: "effect-request.schema.json", value: requestFixture(t)},
		{name: "succeeded receipt", schema: "effect-receipt.schema.json", value: successReceiptFixture(t)},
		{name: "failed receipt", schema: "effect-receipt.schema.json", value: failed},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			schema := compileEffectSchema(t, test.schema)
			document, err := json.Marshal(test.value)
			if err != nil {
				t.Fatal(err)
			}
			decoded, err := jsonschema.UnmarshalJSON(bytes.NewReader(document))
			if err != nil {
				t.Fatal(err)
			}
			if err := schema.Validate(decoded); err != nil {
				t.Fatalf("contract does not satisfy schema: %v", err)
			}
		})
	}
}

func TestReceiptSchemaRejectsUnknownOutcome(t *testing.T) {
	schema := compileEffectSchema(t, "effect-receipt.schema.json")
	document, err := json.Marshal(successReceiptFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(document, &decoded); err != nil {
		t.Fatal(err)
	}
	decoded["status"] = "unknown"
	decoded["code"] = JobOutcomeUnknownCode
	if err := schema.Validate(decoded); err == nil {
		t.Fatal("schema accepted unknown provider outcome receipt")
	}
}

func compileEffectSchema(t *testing.T, name string) *jsonschema.Schema {
	t.Helper()
	compiler := jsonschema.NewCompiler()
	for _, candidate := range []string{
		"effect-request.schema.json", "effect-receipt.schema.json",
	} {
		file, err := os.Open(candidate)
		if err != nil {
			t.Fatal(err)
		}
		value, err := jsonschema.UnmarshalJSON(file)
		file.Close()
		if err != nil {
			t.Fatalf("decode schema %s: %v", candidate, err)
		}
		if err := compiler.AddResource(schemaBase+candidate, value); err != nil {
			t.Fatalf("register schema %s: %v", candidate, err)
		}
	}
	schema, err := compiler.Compile(schemaBase + name)
	if err != nil {
		t.Fatalf("compile schema %s: %v", name, err)
	}
	return schema
}

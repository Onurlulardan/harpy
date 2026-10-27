package main

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

func TestConfigSchemaAcceptsBothExactProfiles(t *testing.T) {
	schema := compileInvocationSchema(t)
	for _, action := range []string{
		"adopt_authority", "issue_finalizer", "authorize_finalization",
		"record_full_rollout",
	} {
		value := validInvocationFixture(t, t.TempDir(), action)
		document, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		decoded, err := jsonschema.UnmarshalJSON(bytes.NewReader(document))
		if err != nil {
			t.Fatal(err)
		}
		if err := schema.Validate(decoded); err != nil {
			t.Fatalf("schema rejected %s: %v", action, err)
		}
	}
}

func TestConfigSchemaRejectsCrossProfile(t *testing.T) {
	schema := compileInvocationSchema(t)
	value := validInvocationFixture(t, t.TempDir(), "adopt_authority")
	value["authorize_finalization"] = map[string]any{}
	document, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := jsonschema.UnmarshalJSON(bytes.NewReader(document))
	if err != nil {
		t.Fatal(err)
	}
	if err := schema.Validate(decoded); err == nil {
		t.Fatal("schema accepted cross-profile fields")
	}
}

func compileInvocationSchema(t *testing.T) *jsonschema.Schema {
	t.Helper()
	file, err := os.Open("cutover-authority-config.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	value, err := jsonschema.UnmarshalJSON(file)
	if err != nil {
		t.Fatal(err)
	}
	compiler := jsonschema.NewCompiler()
	const location = "https://harpy.io/contracts/external-cutover-authority-config.v1.schema.json"
	if err := compiler.AddResource(location, value); err != nil {
		t.Fatal(err)
	}
	schema, err := compiler.Compile(location)
	if err != nil {
		t.Fatal(err)
	}
	return schema
}

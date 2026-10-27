package main

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

func TestFullRolloutEvidenceSchemaAcceptsExactDocument(t *testing.T) {
	document, _ := fullRolloutFixture(t)
	raw, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	schemaDocument, err := os.Open("full-rollout-completion-evidence.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	defer schemaDocument.Close()
	value, err := jsonschema.UnmarshalJSON(schemaDocument)
	if err != nil {
		t.Fatal(err)
	}
	compiler := jsonschema.NewCompiler()
	const location = "https://harpy.io/contracts/full-rollout-completion-evidence.v1.schema.json"
	if err := compiler.AddResource(location, value); err != nil {
		t.Fatal(err)
	}
	schema, err := compiler.Compile(location)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := jsonschema.UnmarshalJSON(bytes.NewReader(raw))
	if err != nil || schema.Validate(decoded) != nil {
		t.Fatalf("completion evidence schema rejected fixture: %v", err)
	}
}

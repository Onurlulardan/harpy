package eventsv1

import (
	"bytes"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

func TestIAMIdentityChangePayloadSchemaIsClosed(t *testing.T) {
	compiler := jsonschema.NewCompiler()
	resource := readEventSchema(t, "iam-identity-change-v1.schema.json")
	identifier := resource.(map[string]any)["$id"].(string)
	if err := compiler.AddResource(identifier, resource); err != nil {
		t.Fatal(err)
	}
	schema, err := compiler.Compile(identifier)
	if err != nil {
		t.Fatal(err)
	}
	valid := []byte(`{
        "clock":11,
        "command":"iam.identity.user.create",
        "resource":"user",
        "result":"iam.identity.user.create.succeeded",
        "revision":5
	}`)
	value, err := jsonschema.UnmarshalJSON(bytes.NewReader(valid))
	validationErr := schema.Validate(value)
	if err != nil || validationErr != nil {
		t.Fatalf("valid identity event payload was rejected: %v/%v", err, validationErr)
	}
	leaking := bytes.Replace(valid, []byte(`"revision":5`),
		[]byte(`"revision":5,"email":"user@example.com"`), 1)
	value, err = jsonschema.UnmarshalJSON(bytes.NewReader(leaking))
	if err != nil || schema.Validate(value) == nil {
		t.Fatal("identity event payload accepted an undeclared confidential field")
	}
}

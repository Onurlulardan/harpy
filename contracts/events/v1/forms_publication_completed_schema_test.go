package eventsv1

import (
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

func TestFormsPublicationCompletedSchemaSealsStateFailureInvariant(t *testing.T) {
	value := readEventSchema(t, "forms-publication-completed-v1.schema.json")
	compiler := jsonschema.NewCompiler()
	identifier := value.(map[string]any)["$id"].(string)
	if err := compiler.AddResource(identifier, value); err != nil {
		t.Fatal(err)
	}
	schema, err := compiler.Compile(identifier)
	if err != nil {
		t.Fatal(err)
	}
	base := map[string]any{
		"definition_id":            "019b0000-0000-7000-8000-000000000001",
		"publication_id":           "019b0000-0000-7000-8000-000000000002",
		"publication_revision":     2,
		"schema_result_generation": 4,
		"schema_job_id":            "019b0000-0000-7000-8000-000000000003",
		"version_id":               "019b0000-0000-7000-8000-000000000004",
	}
	for _, fixture := range []struct {
		state string
		code  any
		valid bool
	}{
		{state: "published", code: nil, valid: true},
		{state: "failed", code: "SCHEMA_JOB_FAILED", valid: true},
		{state: "published", code: "SCHEMA_JOB_FAILED", valid: false},
		{state: "failed", code: nil, valid: false},
	} {
		document := make(map[string]any, len(base)+2)
		for key, item := range base {
			document[key] = item
		}
		document["state"], document["failure_code"] = fixture.state, fixture.code
		if valid := schema.Validate(document) == nil; valid != fixture.valid {
			t.Fatalf("state %s code %v validity = %t", fixture.state, fixture.code, valid)
		}
	}
}

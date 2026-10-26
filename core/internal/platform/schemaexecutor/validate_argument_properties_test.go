package schemaexecutor

import (
	"encoding/json"
	"testing"

	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
)

func TestDecodeArgumentsRequiresBooleanProperties(t *testing.T) {
	tests := []struct {
		name      string
		operation schemav1.OperationID
		arguments string
	}{
		{"column nullable", schemav1.AddFieldColumn,
			`{"field_id":"x","storage_type":"text"}`},
		{"index unique", schemav1.CreateIndex,
			`{"index_id":"x","field_ids":["y"],"method":"btree","active_records_only":false}`},
		{"index active records", schemav1.CreateIndex,
			`{"index_id":"x","field_ids":["y"],"method":"btree","unique":false}`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var arguments schemav1.OperationArguments
			if err := json.Unmarshal([]byte(test.arguments), &arguments); err != nil {
				t.Fatal(err)
			}
			_, err := decodeArguments(schemav1.Step{
				OperationID: test.operation, Arguments: arguments,
			})
			if err == nil {
				t.Fatal("expected missing required property rejection")
			}
		})
	}
}

func TestDecodeArgumentsRejectsNullRequiredBoolean(t *testing.T) {
	var arguments schemav1.OperationArguments
	document := `{"field_id":"x","storage_type":"text","nullable":null}`
	if err := json.Unmarshal([]byte(document), &arguments); err != nil {
		t.Fatal(err)
	}
	_, err := decodeArguments(schemav1.Step{
		OperationID: schemav1.AddFieldColumn, Arguments: arguments,
	})
	if err == nil {
		t.Fatal("expected null required property rejection")
	}
}

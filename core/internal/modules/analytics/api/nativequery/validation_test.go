package nativequery

import (
	"encoding/json"
	"testing"
)

func TestPlanRejectsDuplicateFormSlugs(t *testing.T) {
	query := json.RawMessage(
		`{"select":["id"],"page":{"size":50,"after":null},"count":"none","include":[]}`,
	)
	plan := Plan{
		Contract: ContractV1,
		Sources: []Source{
			{Alias: "employees", FormSlug: "hr-employees", Query: query},
			{Alias: "employees_copy", FormSlug: "hr-employees", Query: query},
		},
		Joins: []Join{{
			Kind: "left", Source: "employees", SourceField: "id",
			Target: "employees_copy", TargetField: "id",
		}},
		Fields: []Field{
			{Alias: "employee_id", Source: "employees", Field: "id"},
		},
		Limit: 50,
	}
	if err := plan.Validate(); err == nil {
		t.Fatal("duplicate native Forms dependency was accepted")
	}
}

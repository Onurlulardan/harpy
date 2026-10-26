package catalogowner

import (
	"encoding/json"
	"testing"

	nativequery "github.com/Onurlulardan/harpy/core/internal/modules/analytics/api/nativequery"
)

func TestNativeJoinSourcesRejectsAmbiguousOrUnknownFilters(t *testing.T) {
	base := endpointDefinition{
		DataSources: []endpointSource{
			{Alias: "employees", FormSlug: "hr-employees"},
			{Alias: "departments", FormSlug: "hr-departments"},
		},
	}
	fields := []nativequery.Field{
		{Alias: "employee_id", Source: "employees", Field: "id"},
		{Alias: "department_id", Source: "departments", Field: "id"},
	}
	for name, field := range map[string]string{
		"unqualified": "status",
		"unknown":     "payroll.status",
	} {
		t.Run(name, func(t *testing.T) {
			definition := base
			definition.Filters = []endpointFilter{{
				Field: field, Operator: "=", Value: json.RawMessage(`"active"`),
			}}
			if _, err := nativeJoinSources(definition, fields); err == nil {
				t.Fatalf("native source filter %q was accepted", field)
			}
		})
	}
}

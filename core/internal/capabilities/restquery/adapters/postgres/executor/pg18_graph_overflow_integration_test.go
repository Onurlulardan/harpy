//go:build integration

package executor

import (
	"errors"
	"testing"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func pg18VerifyGraphOccurrenceLimit(t *testing.T, gate *pg18Gate) {
	pg18ResetGraphTenant(t, gate.ctx, gate.admin, "tenant_a")
	schema := "tenant_a"
	relationA := pg18GraphQualified(schema, pg18GraphRelationTable(graphRelationA))
	relationB := pg18GraphQualified(schema, pg18GraphRelationTable(graphRelationB))
	relationC := pg18GraphQualified(schema, pg18GraphRelationTable(graphRelationC))
	pg18ExecGraphSeed(t, gate.ctx, gate.admin, []pg18GraphSeedStatement{
		{`INSERT INTO tenant_a.frm_graph_root
SELECT ('85000000-0000-4000-8000-'||
pg_catalog.lpad(series.value::text,12,'0'))::uuid,
'Overflow '||pg_catalog.lpad(series.value::text,3,'0')
FROM pg_catalog.generate_series(1,334) AS series(value)`, nil},
		{"INSERT INTO tenant_a.frm_graph_a VALUES($1,'Shared Child')", []any{pg18GraphAOne}},
		{"INSERT INTO tenant_a.frm_graph_b VALUES($1,'Shared Owner')", []any{pg18GraphBOne}},
		{"INSERT INTO tenant_a.frm_graph_c VALUES($1,'Shared Nested')", []any{pg18GraphCOne}},
		{"INSERT INTO " + relationA + `
SELECT fld_graph_root_id,$1,1 FROM tenant_a.frm_graph_root`, []any{pg18GraphAOne}},
		{"INSERT INTO " + relationB + `
SELECT fld_graph_root_id,$1,1 FROM tenant_a.frm_graph_root
ORDER BY fld_graph_root_label LIMIT 333`, []any{pg18GraphBOne}},
		{"INSERT INTO " + relationC + " VALUES($1,$2,1)",
			[]any{pg18GraphAOne, pg18GraphCOne}},
	})
	fixture := pg18GraphPlan(t, pg18TenantA, api.CountModeNone, 500)
	harness := pg18NewGraphHarness(t, gate)
	result, err := pg18ExecuteGraph(t, gate, harness.executor, pg18TenantA, fixture)
	var queryError api.QueryError
	if !errors.As(err, &queryError) ||
		queryError.Code() != restqueryv1.ErrorQueryBudgetExceeded ||
		result.Validate() == nil {
		t.Fatalf("result=%v error=%v code=%s database=%v calls=%v",
			result.Validate(), err, queryError.Code(),
			harness.observation.errorSnapshot(), harness.observation.snapshot())
	}
	issues := queryError.Issues()
	if len(issues) != 1 || issues[0].Path != "" ||
		issues[0].Code != string(restqueryv1.IssueIncludeResultCountExceeded) {
		t.Fatalf("overflow issues=%v", issues)
	}
	if _, pageErr := result.Page(); !errors.Is(pageErr, ErrContract) {
		t.Fatalf("partial page escaped overflow: %v", pageErr)
	}
}

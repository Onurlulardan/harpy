//go:build integration

package executor

import (
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func pg18VerifyGraphEmptyFrontiers(t *testing.T, gate *pg18Gate) {
	pg18ResetGraphTenant(t, gate.ctx, gate.admin, "tenant_a")
	plan := pg18GraphPlan(t, pg18TenantA, api.CountModeNone, 10)
	empty := pg18NewGraphHarness(t, gate)
	result, err := pg18ExecuteGraph(t, gate, empty.executor, pg18TenantA, plan)
	if err != nil || len(pg18Documents(t, result)) != 0 {
		t.Fatalf("empty root result=%v error=%v", result.Validate(), err)
	}
	pg18AssertEmptyFrontierOwners(t, empty)
	pg18AssertGraphTableCalls(t, empty.observation.snapshot(), map[string]int{
		"frm_graph_root": 1,
	})

	if _, err := gate.admin.Exec(gate.ctx,
		"INSERT INTO tenant_a.frm_graph_root VALUES($1,'Edge-free Root')",
		pg18GraphRootOne,
	); err != nil {
		t.Fatal(err)
	}
	edgeFree := pg18NewGraphHarness(t, gate)
	result, err = pg18ExecuteGraph(t, gate, edgeFree.executor, pg18TenantA, plan)
	if err != nil {
		t.Fatal(err)
	}
	documents := pg18Documents(t, result)
	if len(documents) != 1 {
		t.Fatalf("edge-free documents=%v", documents)
	}
	included := pg18GraphObject(t, documents[0]["_included"])
	if len(pg18GraphList(t, included["children"])) != 0 || included["owner"] != nil {
		t.Fatalf("edge-free included=%v", included)
	}
	pg18AssertEmptyFrontierOwners(t, edgeFree)
	pg18AssertGraphTableCalls(t, edgeFree.observation.snapshot(), map[string]int{
		"frm_graph_root":                       1,
		pg18GraphRelationTable(graphRelationA): 1,
		pg18GraphRelationTable(graphRelationB): 1,
	})
}

func pg18AssertEmptyFrontierOwners(t *testing.T, harness pg18GraphHarness) {
	t.Helper()
	if harness.forms.calls.Load() != 2 || harness.crm.calls.Load() != 1 ||
		harness.iam.calls.Load() != 1 {
		t.Fatalf("owners forms/crm/iam=%d/%d/%d",
			harness.forms.calls.Load(), harness.crm.calls.Load(), harness.iam.calls.Load())
	}
}

func pg18AssertGraphTableCalls(
	t *testing.T,
	calls []pg18Call,
	want map[string]int,
) {
	t.Helper()
	tables := []string{
		"frm_graph_root", "frm_graph_a", "frm_graph_b", "frm_graph_c",
		pg18GraphRelationTable(graphRelationA),
		pg18GraphRelationTable(graphRelationB),
		pg18GraphRelationTable(graphRelationC),
	}
	for _, table := range tables {
		explains, data := 0, 0
		for _, call := range calls {
			if !strings.Contains(call.statement, table) {
				continue
			}
			if call.kind == "row" && strings.HasPrefix(call.statement, "EXPLAIN ") {
				explains++
			} else if call.kind == "query" {
				data++
			}
		}
		if explains != want[table] || data != want[table] {
			t.Fatalf("table=%s explain/data=%d/%d want=%d calls=%v",
				table, explains, data, want[table], calls)
		}
	}
}

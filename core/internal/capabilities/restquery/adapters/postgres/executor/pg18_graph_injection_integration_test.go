//go:build integration

package executor

import (
	"strings"
	"testing"
)

func pg18VerifyGraphInjection(t *testing.T, gate *pg18Gate) {
	pg18ResetGraphTenant(t, gate.ctx, gate.admin, "tenant_a")
	pg18SeedGraphHappyA(t, gate)
	payload := `x' OR true; DROP TABLE frm_graph_a; --`
	if _, err := gate.admin.Exec(gate.ctx,
		"INSERT INTO tenant_a.frm_graph_a VALUES($1,$2)",
		pg18GraphASnapshot, payload,
	); err != nil {
		t.Fatal(err)
	}
	relation := pg18GraphQualified("tenant_a", pg18GraphRelationTable(graphRelationA))
	if _, err := gate.admin.Exec(gate.ctx,
		"INSERT INTO "+relation+" VALUES($1,$2,4)",
		pg18GraphRootOne, pg18GraphASnapshot,
	); err != nil {
		t.Fatal(err)
	}
	fixture := pg18GraphInjectionPlan(t, pg18TenantA, payload)
	harness := pg18NewGraphHarness(t, gate)
	result, err := pg18ExecuteGraph(
		t, gate, harness.executor, pg18TenantA, fixture,
	)
	if err != nil {
		t.Fatalf("execute=%v database=%v calls=%v", err,
			harness.observation.errorSnapshot(), harness.observation.snapshot())
	}
	documents := pg18Documents(t, result)
	first := pg18GraphObject(t, documents[0]["_included"])
	children := pg18GraphList(t, first["children"])
	if len(children) != 1 || pg18GraphObject(t, children[0])["label"] != payload {
		t.Fatalf("filtered children=%v", children)
	}
	bound := false
	for _, call := range harness.observation.snapshot() {
		if strings.Contains(call.statement, payload) {
			t.Fatalf("include payload entered SQL text: %s", call.statement)
		}
		for _, argument := range call.arguments {
			if value, ok := argument.(string); ok && value == payload {
				bound = true
			}
		}
	}
	if !bound {
		t.Fatal("include predicate payload was not a bind argument")
	}
	var targetTable, rootTable bool
	if err := gate.admin.QueryRow(gate.ctx, `SELECT
pg_catalog.to_regclass('tenant_a.frm_graph_a') IS NOT NULL,
pg_catalog.to_regclass('tenant_a.frm_graph_root') IS NOT NULL`).Scan(
		&targetTable, &rootTable,
	); err != nil || !targetTable || !rootTable {
		t.Fatalf("tables target/root=%t/%t error=%v", targetTable, rootTable, err)
	}
}

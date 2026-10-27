//go:build integration

package executor

import (
	"errors"
	"testing"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func pg18VerifyGraphCardinality(t *testing.T, gate *pg18Gate) {
	pg18ResetGraphTenant(t, gate.ctx, gate.admin, "tenant_a")
	if _, err := gate.admin.Exec(gate.ctx, `INSERT INTO tenant_a.frm_graph_root
VALUES($1,'Null Owner Root')`, pg18GraphRootOne); err != nil {
		t.Fatal(err)
	}
	table := pg18GraphQualified("tenant_a", pg18GraphRelationTable(graphRelationB))
	if _, err := gate.admin.Exec(gate.ctx,
		"INSERT INTO "+table+" VALUES($1,$2,1)",
		pg18GraphRootOne, pg18GraphBMissing,
	); err != nil {
		t.Fatal(err)
	}
	fixture := pg18GraphPlan(t, pg18TenantA, api.CountModeNone, 10)
	harness := pg18NewGraphHarness(t, gate)
	result, err := pg18ExecuteGraph(t, gate, harness.executor, pg18TenantA, fixture)
	if err != nil {
		t.Fatalf("execute=%v database=%v calls=%v", err,
			harness.observation.errorSnapshot(), harness.observation.snapshot())
	}
	documents := pg18Documents(t, result)
	included := pg18GraphObject(t, documents[0]["_included"])
	if included["owner"] != nil || len(pg18GraphList(t, included["children"])) != 0 {
		t.Fatalf("missing to-one target was not null: %v", included)
	}

	pg18ResetGraphTenant(t, gate.ctx, gate.admin, "tenant_a")
	if _, err := gate.admin.Exec(gate.ctx, "ALTER TABLE "+table+
		" DROP CONSTRAINT graph_rel_b_source, DROP CONSTRAINT graph_rel_b_slot"); err != nil {
		t.Fatal(err)
	}
	defer pg18RestoreToOnePostcondition(t, gate, table)
	pg18ExecGraphSeed(t, gate.ctx, gate.admin, []pg18GraphSeedStatement{
		{"INSERT INTO tenant_a.frm_graph_root VALUES($1,'Broken Owner Root')", []any{pg18GraphRootOne}},
		{`INSERT INTO tenant_a.frm_graph_b VALUES
($1,'Owner One'),($2,'Owner Two')`, []any{pg18GraphBOne, pg18GraphBMissing}},
		{"INSERT INTO " + table + " VALUES($1,$2,1),($1,$3,2)",
			[]any{pg18GraphRootOne, pg18GraphBOne, pg18GraphBMissing}},
	})
	harness = pg18NewGraphHarness(t, gate)
	result, err = pg18ExecuteGraph(t, gate, harness.executor, pg18TenantA, fixture)
	var queryError api.QueryError
	status, found := 0, false
	if errors.As(err, &queryError) {
		status, found = queryError.HTTPStatus()
	}
	if !errors.As(err, &queryError) ||
		queryError.Code() != restqueryv1.ErrorRelationCardinalityViolation ||
		!found || status != 409 || result.Validate() == nil {
		t.Fatalf("result=%v error=%v code=%s status=%d/%t",
			result.Validate(), err, queryError.Code(), status, found)
	}
}

func pg18RestoreToOnePostcondition(
	t *testing.T, gate *pg18Gate, table string,
) {
	t.Helper()
	if _, err := gate.admin.Exec(gate.ctx, "TRUNCATE TABLE "+table); err != nil {
		t.Errorf("truncate corrupted to-one fixture: %v", err)
		return
	}
	if _, err := gate.admin.Exec(gate.ctx, "ALTER TABLE "+table+
		" ADD CONSTRAINT graph_rel_b_slot CHECK(edge_slot = 1),"+
		" ADD CONSTRAINT graph_rel_b_source UNIQUE(source_id)"); err != nil {
		t.Errorf("restore to-one postcondition: %v", err)
	}
}

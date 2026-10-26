//go:build integration

package executor

import (
	"context"
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

const pg18InjectionID = "4eeeeeee-eeee-4eee-8eee-eeeeeeeeeeee"

func (gate *pg18Gate) verifyInjectionBinding(t *testing.T) {
	payload := `x' OR true; DROP TABLE frm_executor_fixture; --`
	if _, err := gate.admin.Exec(gate.ctx, `INSERT INTO tenant_a.frm_executor_fixture
(fld_id,fld_name,fld_name_folded,fld_metadata) VALUES($1,$2,$2,'{"injection":true}')`,
		pg18InjectionID, payload,
	); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = gate.admin.Exec(
			context.Background(),
			"DELETE FROM tenant_a.frm_executor_fixture WHERE fld_id=$1", pg18InjectionID,
		)
	})
	port := pg18NewOwnerPort(t, "forms")
	executor := pg18NewExecutor(t, gate.codec, planner.HardPolicy(), port)
	plan := pg18AuthorizedPlan(t, pg18PlanInput{
		tenant: pg18TenantA, owner: "forms", count: api.CountModeExact, size: 10,
		where: pg18NameFilter(t, payload),
	})
	var result Result
	err := gate.coordinator.InTenant(
		gate.ctx, pg18ScopeRequest(t, pg18TenantA, "forms"),
		func(ctx context.Context, scope transaction.TenantScope) error {
			var err error
			result, err = executor.Execute(ctx, scope, plan)
			return err
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	documents := pg18Documents(t, result)
	total, present, err := result.ExactCount()
	if err != nil || !present || total != 1 || len(documents) != 1 ||
		documents[0]["name"] != payload {
		t.Fatalf("result=%v total=%d present=%t err=%v",
			documents, total, present, err)
	}
	bound := false
	for _, call := range port.observation.snapshot() {
		if strings.Contains(call.statement, payload) {
			t.Fatalf("payload entered SQL text: %s", call.statement)
		}
		for _, argument := range call.arguments {
			if value, ok := argument.(string); ok && value == payload {
				bound = true
			}
		}
	}
	if !bound {
		t.Fatal("payload was not observed as a bind argument")
	}
	var tableA, tableB bool
	var countA, countB int
	if err := gate.admin.QueryRow(gate.ctx, `SELECT
pg_catalog.to_regclass('tenant_a.frm_executor_fixture') IS NOT NULL,
pg_catalog.to_regclass('tenant_b.frm_executor_fixture') IS NOT NULL,
(SELECT count(*) FROM tenant_a.frm_executor_fixture),
(SELECT count(*) FROM tenant_b.frm_executor_fixture)`).Scan(
		&tableA, &tableB, &countA, &countB,
	); err != nil {
		t.Fatal(err)
	}
	if !tableA || !tableB || countA != 4 || countB != 2 {
		t.Fatalf("tables=%t/%t counts=%d/%d", tableA, tableB, countA, countB)
	}
}

//go:build integration

package executor

import (
	"context"
	"net/url"
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"
)

func (gate *pg18Gate) verifyV1Injection(t *testing.T) {
	payload := `x' OR true; DROP TABLE frm_executor_fixture; --`
	identifier := "4ddddddd-dddd-4ddd-8ddd-dddddddddddd"
	_, err := gate.admin.Exec(gate.ctx, `INSERT INTO tenant_a.frm_executor_fixture
(fld_id,fld_name,fld_name_folded,fld_metadata) VALUES($1,$2,$2,'{"v1":true}')`,
		identifier, payload)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = gate.admin.Exec(context.Background(),
			"DELETE FROM tenant_a.frm_executor_fixture WHERE fld_id=$1", identifier)
	})
	port := pg18NewOwnerPort(t, "forms")
	executor := pg18NewExecutor(t, gate.codec, planner.HardPolicy(), port)
	raw := "page=2&limit=1&filters[name]=" + url.QueryEscape(payload)
	accepted := pg18V1Plan(t, raw, pg18TenantA)
	result := gate.pg18ExecuteV1(t, executor, pg18TenantA, accepted)
	if records := pg18ResultIDs(t, result); len(records) != 0 {
		t.Fatalf("filtered page two should be empty: %v", records)
	}
	pg18AssertExactTotal(t, result, 1)
	bound := false
	for _, call := range port.observation.snapshot() {
		if strings.Contains(call.statement, payload) {
			t.Fatal("V1 injection payload entered SQL text")
		}
		for _, argument := range call.arguments {
			if value, ok := argument.(string); ok && value == payload {
				bound = true
			}
		}
	}
	if !bound {
		t.Fatal("V1 injection payload was not bound")
	}
	var tableA, tableB bool
	var countA, countB int
	err = gate.admin.QueryRow(gate.ctx, `SELECT
pg_catalog.to_regclass('tenant_a.frm_executor_fixture') IS NOT NULL,
pg_catalog.to_regclass('tenant_b.frm_executor_fixture') IS NOT NULL,
(SELECT count(*) FROM tenant_a.frm_executor_fixture),
(SELECT count(*) FROM tenant_b.frm_executor_fixture)`).Scan(
		&tableA, &tableB, &countA, &countB)
	if err != nil || !tableA || !tableB || countA != 7 || countB != 2 {
		t.Fatalf("tables=%t/%t counts=%d/%d error=%v", tableA, tableB, countA, countB, err)
	}
}

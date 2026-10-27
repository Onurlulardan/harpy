//go:build integration

package executor

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (gate *pg18Gate) verifyReadIsolation(t *testing.T) {
	port := pg18NewOwnerPort(t, "forms")
	limits := planner.HardLimits()
	limits.InteractiveTimeout = 1500 * time.Millisecond
	limits.ExactCountTimeout = 700 * time.Millisecond
	policy, err := planner.NewPolicy(limits)
	if err != nil {
		t.Fatal(err)
	}
	executor := pg18NewExecutor(t, gate.codec, policy, port)
	tests := []struct {
		tenant string
		count  int64
		first  string
	}{
		{pg18TenantA, 3, "Ada"},
		{pg18TenantB, 2, "Tenant B One"},
	}
	for _, test := range tests {
		var result Result
		var timeout, isolation, readOnly string
		err := gate.coordinator.InTenant(
			gate.ctx, pg18ScopeRequest(t, test.tenant, "forms"),
			func(ctx context.Context, scope transaction.TenantScope) error {
				plan := pg18AuthorizedPlan(t, pg18PlanInput{
					tenant: test.tenant, owner: "forms",
					count: api.CountModeExact, size: 10,
				})
				var executeErr error
				result, executeErr = executor.Execute(ctx, scope, plan)
				if executeErr != nil {
					return executeErr
				}
				bound, bindErr := port.bind(scope)
				if bindErr != nil {
					return bindErr
				}
				return bound.database.QueryRow(ctx, `SELECT
current_setting('statement_timeout'),current_setting('transaction_isolation'),
current_setting('transaction_read_only')`).Scan(&timeout, &isolation, &readOnly)
			},
		)
		if err != nil {
			t.Fatalf("tenant %s execute: %v", test.tenant, err)
		}
		documents := pg18Documents(t, result)
		total, present, err := result.ExactCount()
		if err != nil || !present || total != test.count ||
			len(documents) != int(test.count) || documents[0]["name"] != test.first {
			t.Fatalf("tenant=%s total=%d present=%t records=%v err=%v",
				test.tenant, total, present, documents, err)
		}
		if timeout != "700ms" || isolation != "repeatable read" || readOnly != "on" {
			t.Fatalf("timeout=%s isolation=%s read_only=%s", timeout, isolation, readOnly)
		}
		if test.tenant == pg18TenantA {
			pg18AssertDecodedTypes(t, documents[0])
		}
	}
	if port.observation.opened.Load() != 2 || port.observation.closed.Load() != 2 {
		t.Fatalf("rows opened=%d closed=%d", port.observation.opened.Load(),
			port.observation.closed.Load())
	}
}

func pg18Documents(t *testing.T, result Result) []map[string]any {
	t.Helper()
	page, err := result.Page()
	if err != nil {
		t.Fatal(err)
	}
	records, err := page.Records()
	if err != nil {
		t.Fatal(err)
	}
	documents := make([]map[string]any, len(records))
	for index, record := range records {
		if err := json.Unmarshal(record, &documents[index]); err != nil {
			t.Fatal(err)
		}
	}
	return documents
}

func pg18AssertDecodedTypes(t *testing.T, document map[string]any) {
	t.Helper()
	if document["id"] != "40000000-0000-4000-8000-000000000001" ||
		document["name"] != "Ada" {
		t.Fatalf("UUID/text decode differs: %v", document)
	}
	metadata, exists := document["metadata"]
	if !exists || metadata != nil {
		t.Fatalf("JSON SQL NULL decode differs: %v", document)
	}
}

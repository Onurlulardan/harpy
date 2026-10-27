//go:build integration

package executor

import (
	"context"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (gate *pg18Gate) verifyCursorPageTwo(t *testing.T) {
	port := pg18NewOwnerPort(t, "forms")
	executor := pg18NewExecutor(t, gate.codec, planner.HardPolicy(), port)
	firstPlan := pg18AuthorizedPlan(t, pg18PlanInput{
		tenant: pg18TenantA, owner: "forms", count: api.CountModeExact, size: 2,
	})
	first := gate.pg18ExecutePlan(t, executor, firstPlan)
	firstPage, err := first.Page()
	if err != nil {
		t.Fatal(err)
	}
	cursor, present, err := firstPage.NextCursor()
	if err != nil || !present {
		t.Fatalf("page one cursor present=%t error=%v", present, err)
	}
	secondPlan := pg18AuthorizedPlan(t, pg18PlanInput{
		tenant: pg18TenantA, owner: "forms", count: api.CountModeExact,
		size: 2, after: &cursor,
	})
	second := gate.pg18ExecutePlan(t, executor, secondPlan)
	firstDocuments := pg18Documents(t, first)
	secondDocuments := pg18Documents(t, second)
	wantFirst := []string{
		"40000000-0000-4000-8000-000000000001",
		"40000000-0000-4000-8000-000000000002",
	}
	if len(firstDocuments) != 2 || len(secondDocuments) != 1 ||
		firstDocuments[0]["id"] != wantFirst[0] ||
		firstDocuments[1]["id"] != wantFirst[1] ||
		secondDocuments[0]["id"] != "40000000-0000-4000-8000-000000000003" {
		t.Fatalf("page one=%v page two=%v", firstDocuments, secondDocuments)
	}
	pg18AssertExactTotal(t, first, 3)
	pg18AssertExactTotal(t, second, 3)
	secondPage, _ := second.Page()
	if _, hasMore, err := secondPage.NextCursor(); err != nil || hasMore {
		t.Fatalf("page two has_more=%t error=%v", hasMore, err)
	}
	if port.observation.opened.Load() != 2 || port.observation.closed.Load() != 2 {
		t.Fatalf("cursor rows opened=%d closed=%d",
			port.observation.opened.Load(), port.observation.closed.Load())
	}
}

func (gate *pg18Gate) pg18ExecutePlan(
	t *testing.T, executor Executor, plan resolved.AuthorizedPlan,
) Result {
	t.Helper()
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
	return result
}

func pg18AssertExactTotal(t *testing.T, result Result, want int64) {
	t.Helper()
	total, present, err := result.ExactCount()
	if err != nil || !present || total != want {
		t.Fatalf("exact count=%d present=%t error=%v", total, present, err)
	}
}

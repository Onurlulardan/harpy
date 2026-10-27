//go:build integration

package executor

import (
	"context"
	"errors"
	"testing"
	"time"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (gate *pg18Gate) verifyUnauthorizedOwner(t *testing.T) {
	port := pg18NewOwnerPort(t, "crm")
	executor := pg18NewExecutor(t, gate.codec, planner.HardPolicy(), port)
	plan := pg18AuthorizedPlan(t, pg18PlanInput{
		tenant: pg18TenantA, owner: "crm", count: api.CountModeNone, size: 10,
	})
	err := gate.coordinator.InTenant(
		gate.ctx, pg18ScopeRequest(t, pg18TenantA, "forms"),
		func(ctx context.Context, scope transaction.TenantScope) error {
			_, err := executor.Execute(ctx, scope, plan)
			return err
		},
	)
	if !errors.Is(err, ErrContract) || port.calls.Load() != 0 ||
		len(port.observation.snapshot()) != 0 {
		t.Fatalf("error=%v port=%d SQL=%v", err, port.calls.Load(),
			port.observation.snapshot())
	}
}

func (gate *pg18Gate) verifyPlannerRejection(t *testing.T) {
	port := pg18NewOwnerPort(t, "forms")
	policy, err := planner.NewPolicy(planner.Limits{
		RecordStatementCost: 1, RecordRequestCost: 1,
		ExactCountStatementCost: 1, AggregateStatementCost: 1,
		InteractiveTimeout: time.Second, ExactCountTimeout: time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}
	executor := pg18NewExecutor(t, gate.codec, policy, port)
	plan := pg18AuthorizedPlan(t, pg18PlanInput{
		tenant: pg18TenantA, owner: "forms", count: api.CountModeNone, size: 10,
	})
	err = gate.coordinator.InTenant(
		gate.ctx, pg18ScopeRequest(t, pg18TenantA, "forms"),
		func(ctx context.Context, scope transaction.TenantScope) error {
			_, err := executor.Execute(ctx, scope, plan)
			return err
		},
	)
	pg18AssertQueryCode(t, err, restqueryv1.ErrorQueryBudgetExceeded)
	for _, call := range port.observation.snapshot() {
		if call.kind == "query" {
			t.Fatalf("data SQL ran after planner rejection: %+v", call)
		}
	}
	if port.calls.Load() != 1 {
		t.Fatalf("owner port calls=%d", port.calls.Load())
	}
}

func pg18AssertQueryCode(t *testing.T, err error, code restqueryv1.ErrorCode) {
	t.Helper()
	var queryError api.QueryError
	if !errors.As(err, &queryError) || queryError.Code() != code {
		t.Fatalf("error=%v code=%s", err, queryError.Code())
	}
}

//go:build integration

package executor

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type pg18GraphHarness struct {
	executor        Executor
	observation     *pg18Observation
	forms, crm, iam *pg18OwnerPort
}

func pg18VerifyGraphSnapshot(t *testing.T, gate *pg18Gate) {
	pg18ResetGraphTenant(t, gate.ctx, gate.admin, "tenant_a")
	pg18ResetGraphTenant(t, gate.ctx, gate.admin, "tenant_b")
	pg18SeedGraphHappyA(t, gate)
	pg18SeedGraphHappyB(t, gate)
	planA := pg18GraphPlan(t, pg18TenantA, api.CountModeExact, 10)
	firstHarness := pg18NewGraphHarness(t, gate)
	writer := make(chan error, 1)
	firstHarness.observation.setRowsClosedHook(func() {
		select {
		case writer <- pg18InsertSnapshotChild(gate.ctx, gate.admin):
		default:
		}
	})
	first, isolation, readOnly, err := pg18ExecuteGraphSettings(
		t, gate, firstHarness, planA,
	)
	if err != nil {
		t.Fatalf("execute=%v database=%v calls=%v", err,
			firstHarness.observation.errorSnapshot(), firstHarness.observation.snapshot())
	}
	if isolation != "repeatable read" || readOnly != "on" {
		t.Fatalf("isolation=%s read_only=%s", isolation, readOnly)
	}
	select {
	case writerErr := <-writer:
		if writerErr != nil {
			t.Fatal(writerErr)
		}
	default:
		t.Fatal("external snapshot writer did not run between graph stages")
	}
	pg18AssertGraphDocument(t, first, "Alpha child")
	pg18AssertGraphOwners(t, firstHarness)
	pg18AssertGraphSQLOrder(t, firstHarness.observation.snapshot())

	secondHarness := pg18NewGraphHarness(t, gate)
	second, err := pg18ExecuteGraph(t, gate, secondHarness.executor, pg18TenantA, planA)
	if err != nil {
		t.Fatalf("execute=%v database=%v calls=%v", err,
			secondHarness.observation.errorSnapshot(), secondHarness.observation.snapshot())
	}
	pg18AssertGraphDocument(t, second, "Aardvark Snapshot")

	planB := pg18GraphPlan(t, pg18TenantB, api.CountModeExact, 10)
	tenantBHarness := pg18NewGraphHarness(t, gate)
	tenantB, err := pg18ExecuteGraph(
		t, gate, tenantBHarness.executor, pg18TenantB, planB,
	)
	if err != nil {
		t.Fatalf("execute=%v database=%v calls=%v", err,
			tenantBHarness.observation.errorSnapshot(), tenantBHarness.observation.snapshot())
	}
	pg18AssertTenantBGraph(t, tenantB)

	negativeHarness := pg18NewGraphHarness(t, gate)
	_, err = pg18ExecuteGraph(
		t, gate, negativeHarness.executor, pg18TenantB, planA,
	)
	if !errors.Is(err, ErrContract) ||
		negativeHarness.forms.calls.Load() != 0 ||
		negativeHarness.crm.calls.Load() != 0 || negativeHarness.iam.calls.Load() != 0 ||
		len(negativeHarness.observation.snapshot()) != 0 {
		t.Fatalf("tenant mismatch error=%v owners=%d/%d/%d SQL=%v", err,
			negativeHarness.forms.calls.Load(), negativeHarness.crm.calls.Load(),
			negativeHarness.iam.calls.Load(), negativeHarness.observation.snapshot())
	}
}

func pg18NewGraphHarness(t *testing.T, gate *pg18Gate) pg18GraphHarness {
	t.Helper()
	observation := &pg18Observation{}
	forms := pg18NewOwnerPort(t, "forms")
	crm := pg18NewOwnerPort(t, "crm")
	iam := pg18NewOwnerPort(t, "iam")
	forms.observation, crm.observation, iam.observation = observation, observation, observation
	return pg18GraphHarness{
		executor: pg18NewExecutor(
			t, gate.codec, planner.HardPolicy(), forms, crm, iam,
		),
		observation: observation, forms: forms, crm: crm, iam: iam,
	}
}

func pg18ExecuteGraph(
	t *testing.T, gate *pg18Gate, executor Executor, tenant string,
	fixture graphPlanFixture,
) (Result, error) {
	t.Helper()
	var result Result
	err := gate.coordinator.InTenant(
		gate.ctx, pg18ScopeRequest(t, tenant, "forms", "crm", "iam"),
		func(ctx context.Context, scope transaction.TenantScope) error {
			var executeErr error
			result, executeErr = executor.Execute(ctx, scope, fixture.plan)
			return executeErr
		},
	)
	return result, err
}

func pg18ExecuteGraphSettings(
	t *testing.T, gate *pg18Gate, harness pg18GraphHarness,
	fixture graphPlanFixture,
) (Result, string, string, error) {
	t.Helper()
	var result Result
	var isolation, readOnly string
	err := gate.coordinator.InTenant(
		gate.ctx,
		pg18ScopeRequest(t, pg18TenantA, "forms", "crm", "iam"),
		func(ctx context.Context, scope transaction.TenantScope) error {
			var executeErr error
			result, executeErr = harness.executor.Execute(ctx, scope, fixture.plan)
			if executeErr != nil {
				return executeErr
			}
			bound, bindErr := harness.forms.bind(scope)
			if bindErr != nil {
				return bindErr
			}
			return bound.database.QueryRow(ctx, `SELECT
current_setting('transaction_isolation'),
current_setting('transaction_read_only')`).Scan(&isolation, &readOnly)
		},
	)
	return result, isolation, readOnly, err
}

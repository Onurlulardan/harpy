package executor

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/execution"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/cursor"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type inertPort struct{ calls int }

func (port *inertPort) Run(
	transaction.TenantScope,
	execution.Program,
) error {
	port.calls++
	return nil
}

func TestExecutorBindsCompilerAndPageToSameCodec(t *testing.T) {
	fixture := newBatchFixture(
		t, api.CountModeNone, fixtureTenant, fixtureActor,
	)
	executor, _ := fixtureExecutor(t, fixture.codec, planner.HardPolicy())
	batchValue, err := executor.postgresCompiler.CompileRecordBatch(fixture.plan)
	if err != nil {
		t.Fatal(err)
	}
	batch := consumedFixture(t, batchFixture{batch: batchValue, plan: fixture.plan})
	rows := make([][]any, 3)
	for index := range rows {
		rows[index] = fixtureRecordRow(
			t, batch, "Ada", nil,
			[]string{
				"40000000-0000-4000-8000-000000000001",
				"40000000-0000-4000-8000-000000000002",
				"40000000-0000-4000-8000-000000000003",
			}[index],
		)
	}
	driver, _ := readyDriver(t, batch, rows, 0)
	result, err := runBatch(
		context.Background(), driver, batch, executor.codec, executor.policy,
	)
	if err != nil {
		t.Fatal(err)
	}
	page, _ := result.Page()
	token, present, err := page.NextCursor()
	contextValue, _ := batch.record.CursorContext()
	if err != nil || !present {
		t.Fatal("probe page did not produce a cursor")
	}
	if _, err := fixture.codec.Decode(token, contextValue); err != nil {
		t.Fatal("executor cursor was not encoded by its compiler-bound codec")
	}
}

func TestOwnerScopeAndPlanTransplantsFailClosed(t *testing.T) {
	fixture := newBatchFixture(
		t, api.CountModeNone, fixtureTenant, fixtureActor,
	)
	batch := consumedFixture(t, fixture)
	driver, _ := readyDriver(t, batch, nil, 0)
	driver.owner = "crm"
	if _, err := runBatch(
		context.Background(), driver, batch, fixture.codec, planner.HardPolicy(),
	); !errors.Is(err, ErrContract) || len(driver.events) != 0 {
		t.Fatalf("owner transplant error=%v events=%v", err, driver.events)
	}
	other := newBatchFixture(
		t, api.CountModeNone,
		fixtureTenant, uuid.MustParse("30000000-0000-4000-8000-000000000003"),
	)
	if fixture.batch.ValidatePlan(other.plan) == nil {
		t.Fatal("plan transplant was accepted")
	}
	if _, err := consumeBatch(fixture.batch, other.plan); !errors.Is(err, ErrContract) {
		t.Fatalf("plan-transplanted batch consumption error=%v", err)
	}
	executor, port := fixtureExecutor(t, fixture.codec, planner.HardPolicy())
	if _, err := executor.Execute(
		context.Background(), transaction.TenantScope{}, fixture.plan,
	); !errors.Is(err, ErrContract) || port.calls != 0 {
		t.Fatalf("scope transplant error=%v port calls=%d", err, port.calls)
	}
}

func TestExecutorRejectsSubMillisecondPolicyAtConstruction(t *testing.T) {
	fixture := newBatchFixture(
		t, api.CountModeNone, fixtureTenant, fixtureActor,
	)
	limits := planner.HardLimits()
	limits.InteractiveTimeout = time.Nanosecond
	policy, err := planner.NewPolicy(limits)
	if err != nil {
		t.Fatalf("planner fixture policy rejected early: %v", err)
	}
	registry, _ := fixtureRegistry(t)
	if _, err := New(registry, fixture.codec, policy); !errors.Is(err, ErrContract) {
		t.Fatalf("sub-millisecond executor policy error=%v", err)
	}
}

func fixtureExecutor(
	t *testing.T,
	codec cursor.Codec,
	policy planner.Policy,
) (Executor, *inertPort) {
	t.Helper()
	registry, port := fixtureRegistry(t)
	executor, err := New(registry, codec, policy)
	if err != nil {
		t.Fatal(err)
	}
	return executor, port
}

func fixtureRegistry(t *testing.T) (execution.Registry, *inertPort) {
	t.Helper()
	port := &inertPort{}
	registration, err := execution.NewRegistration("forms", port)
	if err != nil {
		t.Fatal(err)
	}
	registry, err := execution.NewRegistry([]execution.Registration{registration})
	if err != nil {
		t.Fatal(err)
	}
	return registry, port
}

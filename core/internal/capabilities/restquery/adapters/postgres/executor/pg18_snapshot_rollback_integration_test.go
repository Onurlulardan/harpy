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

var pg18RollbackSignal = errors.New("pg18 rollback probe")

func (gate *pg18Gate) verifySnapshotRollback(t *testing.T) {
	port := pg18NewOwnerPort(t, "forms")
	executor := pg18NewExecutor(t, gate.codec, planner.HardPolicy(), port)
	plan := pg18AuthorizedPlan(t, pg18PlanInput{
		tenant: pg18TenantA, owner: "forms", count: api.CountModeNone, size: 10,
	})
	var firstPID, secondPID int32
	var isolation, readOnly, resetTimeout string
	err := gate.coordinator.InTenant(
		gate.ctx, pg18ScopeRequest(t, pg18TenantA, "forms"),
		func(ctx context.Context, scope transaction.TenantScope) error {
			if _, err := executor.Execute(ctx, scope, plan); err != nil {
				return err
			}
			bound, err := port.bind(scope)
			if err != nil {
				return err
			}
			if err := bound.database.QueryRow(ctx, `SELECT pg_backend_pid(),
current_setting('transaction_isolation'),current_setting('transaction_read_only')`).Scan(
				&firstPID, &isolation, &readOnly,
			); err != nil {
				return err
			}
			_, err = bound.database.Execute(ctx, `INSERT INTO frm_executor_fixture
(fld_id,fld_name,fld_name_folded) VALUES
('4fffffff-ffff-4fff-8fff-ffffffffffff','forbidden','forbidden')`)
			if pg18SQLState(err) != "25006" {
				t.Fatalf("read-only mutation error=%v SQLSTATE=%s", err, pg18SQLState(err))
			}
			return pg18RollbackSignal
		},
	)
	if !errors.Is(err, pg18RollbackSignal) || isolation != "repeatable read" ||
		readOnly != "on" {
		t.Fatalf("rollback=%v isolation=%s read_only=%s", err, isolation, readOnly)
	}
	err = gate.coordinator.InTenant(
		gate.ctx, pg18ScopeRequest(t, pg18TenantA, "forms"),
		func(ctx context.Context, scope transaction.TenantScope) error {
			bound, err := port.bind(scope)
			if err != nil {
				return err
			}
			return bound.database.QueryRow(ctx, `SELECT pg_backend_pid(),
current_setting('statement_timeout')`).Scan(&secondPID, &resetTimeout)
		},
	)
	if err != nil || firstPID != secondPID || resetTimeout != "0" {
		t.Fatalf("reuse error=%v pid=%d/%d timeout=%s", err,
			firstPID, secondPID, resetTimeout)
	}
	if port.observation.opened.Load() != 1 || port.observation.closed.Load() != 1 {
		t.Fatalf("row lifecycle opened=%d closed=%d",
			port.observation.opened.Load(), port.observation.closed.Load())
	}
}

func pg18SQLState(err error) string {
	type sqlStateError interface{ SQLState() string }
	var state sqlStateError
	if errors.As(err, &state) {
		return state.SQLState()
	}
	return ""
}

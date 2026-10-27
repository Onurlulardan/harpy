//go:build integration

package executor

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

const pg18SnapshotInsertID = "4ddddddd-dddd-4ddd-8ddd-dddddddddddd"

func (gate *pg18Gate) verifySnapshotExactCount(t *testing.T) {
	port := pg18NewOwnerPort(t, "forms")
	executor := pg18NewExecutor(t, gate.codec, planner.HardPolicy(), port)
	plan := pg18AuthorizedPlan(t, pg18PlanInput{
		tenant: pg18TenantA, owner: "forms", count: api.CountModeExact, size: 10,
	})
	rowsClosed := make(chan struct{})
	writerDone := make(chan error, 1)
	writerContext, cancelWriter := context.WithTimeout(gate.ctx, 30*time.Second)
	defer cancelWriter()
	go func() {
		select {
		case <-rowsClosed:
			_, err := gate.admin.Exec(writerContext, `INSERT INTO
tenant_a.frm_executor_fixture(fld_id,fld_name,fld_name_folded,fld_metadata)
VALUES($1,'Snapshot Writer','snapshot writer','{"snapshot":true}')`,
				pg18SnapshotInsertID,
			)
			writerDone <- err
		case <-writerContext.Done():
			writerDone <- writerContext.Err()
		}
	}()
	var closeOnce sync.Once
	var writerErr error
	port.observation.setRowsClosedHook(func() {
		closeOnce.Do(func() {
			close(rowsClosed)
			writerErr = <-writerDone
		})
	})
	t.Cleanup(func() {
		_, _ = gate.admin.Exec(context.Background(),
			"DELETE FROM tenant_a.frm_executor_fixture WHERE fld_id=$1",
			pg18SnapshotInsertID)
	})
	var first Result
	err := gate.coordinator.InTenant(
		gate.ctx, pg18ScopeRequest(t, pg18TenantA, "forms"),
		func(ctx context.Context, scope transaction.TenantScope) error {
			var err error
			first, err = executor.Execute(ctx, scope, plan)
			return err
		},
	)
	if err != nil || writerErr != nil {
		cancelWriter()
		t.Fatalf("execute=%v writer=%v", err, writerErr)
	}
	pg18AssertPageCount(t, first, 3, 3)
	var second Result
	err = gate.coordinator.InTenant(
		gate.ctx, pg18ScopeRequest(t, pg18TenantA, "forms"),
		func(ctx context.Context, scope transaction.TenantScope) error {
			var err error
			second, err = executor.Execute(ctx, scope, plan)
			return err
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	pg18AssertPageCount(t, second, 4, 4)
}

func pg18AssertPageCount(t *testing.T, result Result, records int, total int64) {
	t.Helper()
	documents := pg18Documents(t, result)
	actual, present, err := result.ExactCount()
	if err != nil || !present || len(documents) != records || actual != total {
		t.Fatalf("records=%d/%d count=%d/%d present=%t error=%v",
			len(documents), records, actual, total, present, err)
	}
}

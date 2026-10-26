//go:build integration

package transaction

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
)

const queuedRowID = "aaaaaaaa-8888-4888-8888-888888888888"

func assertQueuedExecutorRejectedAtSeal(
	t *testing.T,
	ctx context.Context,
	coordinator *Coordinator,
	binder postgres.OwnerBinder[scopeProbe],
	database scopeTestDatabase,
) {
	t.Helper()
	request := scopeRequestFor(t, tenantA, 1)
	queuedResult := make(chan error, 1)
	err := coordinator.InTenant(ctx, request, func(ctx context.Context, scope TenantScope) error {
		probe, err := Bind(scope, binder)
		if err != nil {
			return err
		}
		rows, err := probe.database.Query(ctx, "SELECT generate_series(1, 1000000)")
		if err != nil {
			return err
		}
		_ = rows
		started := make(chan struct{})
		go func() {
			close(started)
			_, executeErr := probe.database.Execute(
				context.Background(), "INSERT INTO iam_probe(id,value) VALUES($1,$2)",
				queuedRowID, "must-never-run",
			)
			queuedResult <- executeErr
		}()
		<-started
		select {
		case executeErr := <-queuedResult:
			return fmt.Errorf("queued SQL completed before seal: %v", executeErr)
		case <-time.After(100 * time.Millisecond):
			return nil
		}
	})
	if !errors.Is(err, ErrScopeDatabase) {
		t.Fatalf("queued executor scope error = %v", err)
	}
	select {
	case executeErr := <-queuedResult:
		if !errors.Is(executeErr, postgres.ErrTransactionClosed) {
			t.Fatalf("queued executor error = %v", executeErr)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("queued executor was not released by seal")
	}
	assertScopeRowAbsent(t, ctx, database, queuedRowID)
}

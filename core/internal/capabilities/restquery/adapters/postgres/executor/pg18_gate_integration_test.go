//go:build integration

package executor

import (
	"context"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/cursor"
	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/jackc/pgx/v5"
)

type pg18Gate struct {
	ctx         context.Context
	admin       *pgx.Conn
	coordinator *transaction.Coordinator
	codec       cursor.Codec
}

func TestPG18RestQueryExecutorGate(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
	defer cancel()
	database := pg18StartDatabase(t, ctx)
	pool, codec := pg18OpenRuntime(t, ctx, database.runtimeDSN)
	release, err := postgres.ReserveIntegrationConnection(ctx, pool)
	if err != nil {
		t.Fatal(err)
	}
	defer release()
	coordinator, err := transaction.NewCoordinator(pool)
	if err != nil {
		t.Fatal(err)
	}
	gate := &pg18Gate{
		ctx: ctx, admin: database.admin, coordinator: coordinator, codec: codec,
	}
	t.Run("decode count timeout and tenant isolation", gate.verifyReadIsolation)
	t.Run("unauthorized owner before port", gate.verifyUnauthorizedOwner)
	t.Run("planner rejection before data SQL", gate.verifyPlannerRejection)
	t.Run("record and exact count share snapshot", gate.verifySnapshotExactCount)
	t.Run("read-only rollback and row lifetime", gate.verifySnapshotRollback)
	t.Run("bound injection payload", gate.verifyInjectionBinding)
	t.Run("stable cursor and root exact count", gate.verifyCursorPageTwo)
	t.Run("aggregate rounding overflow and empty", gate.verifyAggregateExecution)
	t.Run("V1 predecessor pagination", gate.verifyV1Execution)
	t.Run("include graph lifecycle and storage", gate.verifyGraphExecution)
}

//go:build integration

package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/schemaexecutor"
)

func runSchemaExecutorVerifyTenantTransportError(
	t *testing.T,
	ctx context.Context,
	environment *schemaExecutorIntegrationEnvironment,
	tenantID string,
) {
	t.Helper()
	holder, err := environment.pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = holder.Rollback(ctx) }()
	var lockedTenant string
	if err := holder.QueryRow(ctx, `
SELECT id::text FROM platform.tenants WHERE id = $1 FOR UPDATE`, tenantID,
	).Scan(&lockedTenant); err != nil {
		t.Fatal(err)
	}
	connection, err := environment.pool.Acquire(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Release()
	session := &schemaExecutorTenantSession{
		connection: connection, tenantID: tenantID,
		schemaName: "tenant_a", cleanup: time.Second,
	}
	verifyCtx, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
	verifyErr := session.verifyTenant(verifyCtx)
	cancel()
	classified := schemaExecutorClassifyError(verifyErr)
	if !errors.Is(classified, context.DeadlineExceeded) {
		t.Fatalf("tenant verification transport error = %v", classified)
	}
	schemaExecutorAssertFault(t, classified,
		schemaexecutor.FaultTransient, "schema_operation_interrupted")
}

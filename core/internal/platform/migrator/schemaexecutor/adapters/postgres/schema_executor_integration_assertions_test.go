//go:build integration

package postgres

import (
	"context"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/schemaexecutor"
)

const schemaExecutorIntegrationDetailSHA256 = "0000000000000000000000000000000000000000000000000000000000000000"

func assertSchemaExecutorRelationExists(
	t *testing.T,
	ctx context.Context,
	environment *schemaExecutorIntegrationEnvironment,
	schemaName string,
	relationName string,
	want bool,
) {
	t.Helper()
	var actual *string
	qualified := pgIdentifier(schemaName) + "." + pgIdentifier(relationName)
	if err := environment.pool.QueryRow(ctx,
		`SELECT pg_catalog.to_regclass($1)::text`, qualified,
	).Scan(&actual); err != nil {
		t.Fatal(err)
	}
	if (actual != nil) != want {
		t.Fatalf("relation %s exists=%t, want %t", qualified, actual != nil, want)
	}
}

func schemaExecutorAssertFault(
	t *testing.T,
	err error,
	wantKind schemaexecutor.FaultKind,
	wantCode string,
) {
	t.Helper()
	if err == nil {
		t.Fatalf("fault = nil, want %s/%s", wantKind, wantCode)
	}
	kind, code := schemaexecutor.FaultDetails(err)
	if kind != wantKind || code != wantCode {
		t.Fatalf("fault = %s/%s (%v), want %s/%s",
			kind, code, err, wantKind, wantCode)
	}
}

func recordSchemaExecutorBlocked(
	t *testing.T,
	ctx context.Context,
	environment *schemaExecutorIntegrationEnvironment,
	lease schemaexecutor.Lease,
	code string,
) {
	t.Helper()
	if err := environment.store.RecordFailure(ctx, schemaexecutor.FailureRequest{
		Lease: lease, State: schemaexecutor.FailureBlocked,
		Code: code, RedactedDetailSHA256: schemaExecutorIntegrationDetailSHA256,
	}); err != nil {
		t.Fatalf("record blocked schema job: %v", err)
	}
}

func assertSchemaExecutorNoActiveJobs(
	t *testing.T,
	ctx context.Context,
	environment *schemaExecutorIntegrationEnvironment,
) {
	t.Helper()
	var count int
	if err := environment.pool.QueryRow(ctx, `
SELECT pg_catalog.count(*)
FROM platform.schema_jobs
WHERE state IN ('queued', 'leased', 'running', 'lease_expired', 'awaiting_runtime')`,
	).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("active schema executor jobs after suite = %d", count)
	}
}

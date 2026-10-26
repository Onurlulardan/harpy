//go:build integration

package postgres

import (
	"context"
	"testing"

	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
)

func assertV2OversizedProgressUnchanged(
	t *testing.T, ctx context.Context,
	environment *schemaExecutorIntegrationEnvironment, plan schemav2.Plan, jobID string,
) {
	t.Helper()
	var rows int64
	var checkpoint, digest *string
	var canonicalBytes *int64
	err := environment.pool.QueryRow(ctx, `SELECT processed_rows,
checkpoint_record_id::text,checkpoint_canonical_bytes,checkpoint_page_sha256
FROM platform.schema_job_steps WHERE job_id=$1 AND ordinal=0`, jobID).Scan(
		&rows, &checkpoint, &canonicalBytes, &digest)
	if err != nil || rows != 0 || checkpoint != nil ||
		canonicalBytes != nil || digest != nil {
		t.Fatalf("oversized progress rows=%d checkpoint=%v bytes=%v digest=%v err=%v",
			rows, checkpoint, canonicalBytes, digest, err)
	}
	mapping := plan.Steps[0].IdentifierMapping
	table := pgIdentifier("tenant_backfill_v2") + "." +
		pgIdentifier(mapping.Form.PhysicalName)
	target := pgIdentifier(mapping.Fields[1].PhysicalName)
	var records, written int
	query := "SELECT pg_catalog.count(*),pg_catalog.count(" + target + ") FROM " + table
	if err := environment.pool.QueryRow(ctx, query).Scan(&records, &written); err != nil ||
		records != 2 || written != 0 {
		t.Fatalf("oversized records=%d written=%d err=%v", records, written, err)
	}
}

//go:build integration

package postgres

import (
	"context"
	"testing"

	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
)

func assertV2BackfillPageDigest(
	t *testing.T, ctx context.Context,
	environment *schemaExecutorIntegrationEnvironment, plan schemav2.Plan,
	after *string, checkpoint string, expectedBytes int64, expectedSHA string,
) {
	t.Helper()
	mapping := plan.Steps[0].IdentifierMapping
	table := pgIdentifier("tenant_backfill_v2") + "." +
		pgIdentifier(mapping.Form.PhysicalName)
	source := pgIdentifier(mapping.Fields[0].PhysicalName)
	rowPayload := "pg_catalog.convert_to(pg_catalog.jsonb_build_array(" +
		"id::text,row_version-1,pg_catalog.to_jsonb(" + source + "))::text,'UTF8')"
	query := "WITH page AS (SELECT COALESCE(pg_catalog.string_agg(" + rowPayload +
		",pg_catalog.convert_to(E'\\n','UTF8') ORDER BY id),''::bytea) AS payload " +
		"FROM " + table + " WHERE ($1::uuid IS NULL OR id>$1::uuid) " +
		"AND id<=$2::uuid) SELECT pg_catalog.octet_length(payload)::bigint," +
		"pg_catalog.encode(pg_catalog.sha256(payload),'hex') FROM page"
	var actualBytes int64
	var actualSHA string
	if err := environment.pool.QueryRow(ctx, query, after, checkpoint).Scan(
		&actualBytes, &actualSHA); err != nil {
		t.Fatal(err)
	}
	if actualBytes != expectedBytes || actualSHA != expectedSHA {
		t.Fatalf("persisted page bytes/digest=%d/%s actual=%d/%s",
			expectedBytes, expectedSHA, actualBytes, actualSHA)
	}
}

func assertV2BackfillComplete(
	t *testing.T, ctx context.Context,
	environment *schemaExecutorIntegrationEnvironment, plan schemav2.Plan, jobID string,
) {
	t.Helper()
	mapping := plan.Steps[0].IdentifierMapping
	table := pgIdentifier("tenant_backfill_v2") + "." +
		pgIdentifier(mapping.Form.PhysicalName)
	source := pgIdentifier(mapping.Fields[0].PhysicalName)
	target := pgIdentifier(mapping.Fields[1].PhysicalName)
	var mismatches int
	if err := environment.pool.QueryRow(ctx, "SELECT count(*) FROM "+table+
		" WHERE "+source+" IS DISTINCT FROM "+target).Scan(&mismatches); err != nil {
		t.Fatal(err)
	}
	var rows int64
	var state string
	var bytes int64
	var digest string
	err := environment.pool.QueryRow(ctx, `SELECT processed_rows,state,
checkpoint_canonical_bytes,checkpoint_page_sha256
FROM platform.schema_job_steps WHERE job_id=$1 AND ordinal=0`, jobID).Scan(
		&rows, &state, &bytes, &digest)
	if err != nil || mismatches != 0 || rows != 1010 || state != "succeeded" ||
		bytes < 1 || !schemaExecutorPageDigest(digest) {
		t.Fatalf("V2 backfill final mismatches=%d rows=%d state=%s bytes=%d err=%v",
			mismatches, rows, state, bytes, err)
	}
}

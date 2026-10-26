//go:build integration

package integrationtest

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

type recoveryEvidenceSnapshot struct {
	plan, mappings, pages, checkpoint, failures, receipts string
}

func readRecoveryEvidence(
	t *testing.T, ctx context.Context, db publicationDatabase, planID uuid.UUID,
) recoveryEvidenceSnapshot {
	t.Helper()
	const query = `SELECT
encode(pg_catalog.sha256(pg_catalog.convert_to(to_jsonb(plan)::text,'UTF8')),'hex'),
encode(pg_catalog.sha256(pg_catalog.convert_to(COALESCE((
 SELECT jsonb_agg(to_jsonb(mapping) ORDER BY mapping.ordinal)
 FROM t_profile_publication.iam_profile_backfill_mappings AS mapping
 WHERE mapping.plan_id=plan.id),'[]'::jsonb)::text,'UTF8')),'hex'),
encode(pg_catalog.sha256(pg_catalog.convert_to(COALESCE((
 SELECT jsonb_agg(to_jsonb(page) ORDER BY page.page_number)
 FROM t_profile_publication.iam_profile_backfill_pages AS page
 WHERE page.plan_id=plan.id),'[]'::jsonb)::text,'UTF8')),'hex'),
encode(pg_catalog.sha256(pg_catalog.convert_to(COALESCE((
 SELECT to_jsonb(checkpoint)
 FROM t_profile_publication.iam_profile_backfill_checkpoints AS checkpoint
 WHERE checkpoint.plan_id=plan.id),'{}'::jsonb)::text,'UTF8')),'hex'),
encode(pg_catalog.sha256(pg_catalog.convert_to(COALESCE((
 SELECT jsonb_agg(to_jsonb(failure) ORDER BY failure.page_number,failure.resource_id)
 FROM t_profile_publication.iam_profile_backfill_errors AS failure
 WHERE failure.plan_id=plan.id),'[]'::jsonb)::text,'UTF8')),'hex'),
encode(pg_catalog.sha256(pg_catalog.convert_to(COALESCE((
 SELECT jsonb_agg(receipt ORDER BY receipt.kind,receipt.identity)
 FROM (
  SELECT 'authorization' AS kind,operation.operation_id::text AS identity,
         to_jsonb(operation) AS evidence
  FROM t_profile_publication.iam_authorization_operation_receipts AS operation
  WHERE operation.result_id=plan.id
  UNION ALL
  SELECT 'maintenance',maintenance.job_id::text,to_jsonb(maintenance)
  FROM t_profile_publication.iam_maintenance_receipts AS maintenance
  JOIN t_profile_publication.iam_profile_backfill_pages AS page
    ON page.job_id=maintenance.job_id WHERE page.plan_id=plan.id
 ) AS receipt),'[]'::jsonb)::text,'UTF8')),'hex')
FROM t_profile_publication.iam_profile_backfill_plans AS plan WHERE plan.id=$1`
	var snapshot recoveryEvidenceSnapshot
	err := db.admin.QueryRow(ctx, query, planID).Scan(
		&snapshot.plan, &snapshot.mappings, &snapshot.pages,
		&snapshot.checkpoint, &snapshot.failures, &snapshot.receipts,
	)
	if err != nil {
		t.Fatalf("read recovery parent evidence: %v", err)
	}
	return snapshot
}

func assertRecoveryEvidenceUnchanged(
	t *testing.T, ctx context.Context, db publicationDatabase,
	planID uuid.UUID, before recoveryEvidenceSnapshot,
) {
	t.Helper()
	after := readRecoveryEvidence(t, ctx, db, planID)
	if after != before {
		t.Fatalf("recovery parent evidence changed\nbefore=%#v\nafter=%#v", before, after)
	}
}

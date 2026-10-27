//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
)

func completeDocumentsCopyEffect(t *testing.T, ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	fixtures := []struct {
		statement string
		arguments []any
	}{
		{`UPDATE documents_adopted_a.documents_blob_copy_checkpoints
SET target_object_key='tenants/evidence/documents/item/version/content',
 target_object_version='object-version-1',target_object_generation=1,
 target_ciphertext_bytes=27,target_plaintext_sha256=source_sha256,
 target_ciphertext_sha256=pg_catalog.decode(repeat('d',64),'hex'),
 target_key_id='documents-kek-1',
 target_wrapped_dek=pg_catalog.decode(repeat('ab',115),'hex'),
 metadata_sha256=source_metadata_sha256,applied_lsn=source_lsn,
 lifecycle_state='VERIFIED',verified_at=pg_catalog.clock_timestamp(),
 updated_at=pg_catalog.clock_timestamp()
WHERE run_id=$1 AND source_version_id=$2`,
			[]any{documentsAdoptionRun, documentsSourceVersion}},
		{`WITH observed AS (SELECT pg_catalog.clock_timestamp() AS at)
UPDATE platform.jobs SET state='succeeded',attempts=1,lease_generation=1,
 updated_at=observed.at,completed_at=observed.at,
 retain_until=observed.at+interval '91 days'
FROM observed WHERE id=$1`, []any{documentsCopyJob}},
		{`INSERT INTO platform.idempotency_receipts (
 receipt_id,record_kind,tenant_id,scope_kind,scope_id,authority_group_id,
 authority_group_generation,authority_domain,cutover_generation,writer_origin,
 request_sha256,retention_class,effect_id,effect_kind,effect_revision,effect_key,
 provider,operation,correlation_id,causation_id,effect_status,effect_code,
 provider_reference,effect_result_ref,effect_result_sha256,completed_at,retain_until)
VALUES ($1,'effect',$2,'tenant',$2,'documents',13,'documents',13,'v2',
 repeat('4',64),'external_effect',$3,'documents.adoption.copy',1,
 'documents:adoption-copy:'||$4::text||':'||$5::text,
 'documents.adoption','copy','documents-copy-1',$6,'succeeded',
 'DOCUMENT_COPY_SUCCEEDED','object-version-1','documents/effect/1',repeat('5',64),
 pg_catalog.clock_timestamp(),pg_catalog.clock_timestamp()+interval '365 days')`,
			[]any{"018f0000-0000-7000-8000-000000000040",
				documentsCutoverTenantID,
				"018f0000-0000-7000-8000-000000000041", documentsAdoptionRun,
				documentsSourceVersion, documentsCopyJob}},
		{`INSERT INTO platform.job_attempts (
 attempt_id,job_id,scope_kind,scope_id,attempt,lease_generation,lease_owner_id,
 started_at,finished_at,outcome,effect_started_at,effect_id,
 effect_request_sha256,effect_receipt_id)
VALUES ($1,$2,'tenant',$3,1,1,$4,
 pg_catalog.clock_timestamp()-interval '1 second',pg_catalog.clock_timestamp(),
 'succeeded',pg_catalog.clock_timestamp()-interval '500 milliseconds',$5,
 repeat('4',64),$6)`, []any{
			"018f0000-0000-7000-8000-000000000042", documentsCopyJob,
			documentsCutoverTenantID, "018f0000-0000-7000-8000-000000000043",
			"018f0000-0000-7000-8000-000000000041",
			"018f0000-0000-7000-8000-000000000040",
		}},
		{`UPDATE platform.cutover_operations SET state='quiescing',
 updated_at=pg_catalog.clock_timestamp()
WHERE operation_id=$1 AND run_id=$2`,
			[]any{documentsCutoverOperation, documentsCutoverRun}},
	}
	for _, fixture := range fixtures {
		if _, err = tx.ExecContext(ctx, fixture.statement,
			fixture.arguments...); err != nil {
			t.Fatal(err)
		}
	}
	if err = tx.Commit(); err != nil {
		t.Fatal(err)
	}
}

func assertDocumentsReceiptImmutable(t *testing.T, ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	_, err := database.ExecContext(ctx, `UPDATE
 documents_adopted_a.documents_v1_cutover_receipts
SET receipt_sha256=repeat('f',64)`)
	requireDocumentsPGError(t, err, "55000")
}

func switchDocumentsAuthorityToV2(t *testing.T, ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(ctx, `UPDATE platform.domain_cutovers
SET state='v2',writer='v2' WHERE scope_kind='tenant' AND scope_id=$1
 AND domain='documents'`,
		documentsCutoverTenantID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = tx.ExecContext(ctx, `UPDATE platform.cutover_groups
SET state='v2',writer='v2'
WHERE scope_kind='tenant' AND scope_id=$1 AND group_id='documents'`,
		documentsCutoverTenantID)
	if err != nil {
		t.Fatal(err)
	}
	if err = tx.Commit(); err != nil {
		t.Fatal(err)
	}
}

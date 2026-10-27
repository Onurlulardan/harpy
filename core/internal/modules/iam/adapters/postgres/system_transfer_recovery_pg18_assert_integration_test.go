//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

const seedIAMExpiredReconciliationExport = `
INSERT INTO iam_system_record_exports(
 export_id,tenant_id,job_id,idempotency_receipt_id,resource_kind,actor_id,
 token_version,authorization_version,authorization_clock,catalog_version_id,
 catalog_sha256,query_document,query_sha256,request_sha256,
 idempotency_key_sha256,format,state,header_sha256,ciphertext_sha256,
 ciphertext_bytes,key_id,wrapped_dek,wrap_nonce,attempt_generation,
 failure_code,effect_id,effect_lease_owner_id,effect_object_key,
 effect_request_sha256,effect_started_at,outcome_receipt_id,created_at,
 started_at,completed_at,metadata_retain_until,reconciliation_job_id,
 reconciliation_state,reconciliation_attempt_generation,
 reconciliation_effect_id,reconciliation_receipt_id,
 reconciliation_request_sha256,reconciliation_started_at
) SELECT
 $1::uuid,$2::uuid,$3::uuid,$4::uuid,
 'users','91000000-0000-4000-8000-000000000001',
 1,1,1,'019be100-0000-7000-8000-000000000701',
 decode(repeat('51',32),'hex'),convert_to('{}','UTF8'),
 sha256(convert_to('{}','UTF8')),decode(repeat($10,32),'hex'),
 decode(repeat($10,32),'hex'),'csv','outcome_unknown',
 decode(repeat('52',32),'hex'),decode(repeat('53',32),'hex'),128,
 'iam-recovery-fixture',decode(repeat('54',32),'hex'),
 decode(repeat('55',12),'hex'),1,'JOB_OUTCOME_UNKNOWN',$5::uuid,$6::uuid,
 'iam-system-exports/'||$1::uuid::text,decode(repeat('56',32),'hex'),
 clock_timestamp()-interval '10 minutes',$7::uuid,
 clock_timestamp()-interval '12 minutes',
 clock_timestamp()-interval '11 minutes',
 clock_timestamp()-interval '9 minutes',clock_timestamp()+interval '400 days',
 $8::uuid,'running',$11::bigint,$9::uuid,$12::uuid,
 decode(repeat($10,32),'hex'),
 attempt.effect_started_at
FROM platform.job_attempts AS attempt WHERE attempt.attempt_id=$13::uuid;`

func assertIAMExportReconciliationExpiredRecovery(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	seedIAMExpiredReconciliationPlatform(t, ctx, database)
	for _, fixture := range iamExpiredReconciliationFixtures {
		mustExecIAMTenant(t, ctx, database, iamFoundationTenants[0].schema,
			seedIAMExpiredReconciliationExport,
			fixture.exportID, iamFoundationTenants[0].id, fixture.originalJobID,
			fixture.originalKeyID, fixture.effectID, fixture.leaseOwnerID,
			fixture.outcomeID, fixture.jobID, fixture.effectID, fixture.marker,
			fixture.generation, fixture.receiptID, fixture.attemptID)
	}
	runtime := openIAMExpiredReconciliationRuntime(t, ctx, database)
	defer runtime.Close()
	assertIAMExpiredRecoveryCall(t, ctx, runtime,
		iamFoundationTenants[1], false, "")
	assertIAMExpiredRecoveryCall(t, ctx, runtime,
		iamFoundationTenants[0], true, "requeued")
	advanceIAMExpiredRecoveryAuthority(t, ctx, database)
	assertIAMExpiredRecoveryCall(t, ctx, runtime,
		iamFoundationTenants[0], true, "parked")
	assertIAMExpiredRecoveryCall(t, ctx, runtime,
		iamFoundationTenants[0], false, "")
	assertIAMExpiredRecoveryRows(t, ctx, database)
}

func openIAMExpiredReconciliationRuntime(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) *sql.Conn {
	t.Helper()
	connection, err := database.Conn(ctx)
	if err != nil {
		t.Fatalf("reserve IAM recovery runtime connection: %v", err)
	}
	if _, err = connection.ExecContext(ctx,
		"SET SESSION AUTHORIZATION harpy_v2_runtime"); err != nil {
		connection.Close()
		t.Fatalf("bind IAM recovery runtime session: %v", err)
	}
	if _, err = connection.ExecContext(ctx, "SET ROLE harpy_runtime"); err != nil {
		connection.Close()
		t.Fatalf("assume IAM recovery runtime role: %v", err)
	}
	return connection
}

func assertIAMExpiredRecoveryCall(
	t *testing.T,
	ctx context.Context,
	connection *sql.Conn,
	tenant iamFoundationTenant,
	wantRecovered bool,
	wantResolution string,
) {
	t.Helper()
	query := `SELECT recovered,resolution FROM ` + tenant.schema +
		`.iam_system_record_export_reconcile_expired_v1($1,$2)`
	var recovered bool
	var resolution sql.NullString
	if err := connection.QueryRowContext(ctx, query, tenant.schema, tenant.id).
		Scan(&recovered, &resolution); err != nil {
		t.Fatalf("recover expired IAM export for %s: %v", tenant.schema, err)
	}
	if recovered != wantRecovered || resolution.Valid != (wantResolution != "") ||
		resolution.String != wantResolution {
		t.Fatalf("expired IAM recovery for %s = %v/%q, want %v/%q",
			tenant.schema, recovered, resolution.String,
			wantRecovered, wantResolution)
	}
}

func advanceIAMExpiredRecoveryAuthority(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	statements := []string{
		`UPDATE platform.cutover_groups SET generation=2
WHERE scope_kind='tenant' AND scope_id=$1 AND group_id='iam_system_resources'`,
		`UPDATE platform.domain_cutovers SET generation=2
WHERE scope_kind='tenant' AND scope_id=$1 AND domain='iam'`,
	}
	for _, statement := range statements {
		if _, err := database.ExecContext(
			ctx, statement, iamFoundationTenants[0].id); err != nil {
			t.Fatalf("advance IAM recovery authority: %v", err)
		}
	}
}

func assertIAMExpiredRecoveryRows(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	connection, transaction := beginIAMTenantTx(
		t, ctx, database, iamFoundationTenants[0].schema)
	defer connection.Close()
	defer transaction.Rollback()
	var requeued, parked int
	err := transaction.QueryRowContext(ctx, `SELECT
 count(*) FILTER(WHERE reconciliation_state='retry_wait'
   AND reconciliation_code='IAM_SYSTEM_EXPORT_ARTIFACT_RECONCILIATION_UNKNOWN'
   AND reconciliation_next_at IS NOT NULL),
 count(*) FILTER(WHERE reconciliation_state='parked'
   AND reconciliation_code='IAM_SYSTEM_EXPORT_ARTIFACT_RECONCILIATION_PARKED'
   AND reconciliation_next_at IS NULL)
FROM iam_system_record_exports WHERE export_id IN ($1,$2)`,
		iamExpiredReconciliationFixtures[0].exportID,
		iamExpiredReconciliationFixtures[1].exportID).Scan(&requeued, &parked)
	if err != nil || requeued != 1 || parked != 1 {
		t.Fatalf("IAM recovery tenant outcomes=%d/%d: %v", requeued, parked, err)
	}
	var jobs, attempts, receipts int
	err = database.QueryRowContext(ctx, `SELECT
 (SELECT count(*) FROM platform.jobs WHERE id IN ($1,$2)
   AND state IN ('queued','dead')),
 (SELECT count(*) FROM platform.job_attempts WHERE job_id IN ($1,$2)
   AND outcome IN ('requeued','dead')),
 (SELECT count(*) FROM platform.idempotency_receipts
   WHERE receipt_id IN ($3,$4) AND effect_status='failed')`,
		iamExpiredReconciliationFixtures[0].jobID,
		iamExpiredReconciliationFixtures[1].jobID,
		iamExpiredReconciliationFixtures[0].receiptID,
		iamExpiredReconciliationFixtures[1].receiptID).
		Scan(&jobs, &attempts, &receipts)
	if err != nil || jobs != 2 || attempts != 2 || receipts != 2 {
		t.Fatalf("IAM recovery platform outcomes=%d/%d/%d: %v",
			jobs, attempts, receipts, err)
	}
}

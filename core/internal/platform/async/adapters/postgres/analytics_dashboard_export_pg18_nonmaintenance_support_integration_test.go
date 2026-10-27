//go:build integration

package postgres

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	effectsv1 "github.com/Onurlulardan/harpy/contracts/effects/v1"
	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const nonMaintenanceCutoverReceiptGuardPG18 = `
CREATE FUNCTION platform.platform_async_assert_cutover_effect_receipt_v1(
    p_receipt platform.idempotency_receipts
) RETURNS void LANGUAGE plpgsql SECURITY DEFINER SET search_path=pg_catalog
AS $function$
DECLARE v_job platform.jobs%ROWTYPE;
BEGIN
    SELECT job.* INTO v_job FROM platform.jobs AS job
    JOIN platform.cutover_job_bindings AS binding ON binding.job_id=job.id
    WHERE job.id=p_receipt.causation_id FOR SHARE OF job,binding;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='P5603',
            MESSAGE='cutover effect receipt job is missing';
    END IF;
    PERFORM platform.platform_async_assert_job_row_authority_v1(v_job);
    IF p_receipt.record_kind IS DISTINCT FROM 'effect' OR
       p_receipt.tenant_id IS DISTINCT FROM v_job.tenant_id OR
       p_receipt.scope_kind IS DISTINCT FROM v_job.scope_kind OR
       p_receipt.scope_id IS DISTINCT FROM v_job.scope_id OR
       p_receipt.authority_group_id IS DISTINCT FROM v_job.authority_group_id OR
       p_receipt.authority_group_generation IS DISTINCT FROM
           v_job.authority_group_generation OR
       p_receipt.authority_domain IS DISTINCT FROM v_job.authority_domain OR
       p_receipt.cutover_generation IS DISTINCT FROM v_job.cutover_generation OR
       p_receipt.writer_origin IS DISTINCT FROM 'v2' OR
       p_receipt.effect_kind IS DISTINCT FROM v_job.kind OR
       p_receipt.effect_revision IS DISTINCT FROM v_job.payload_revision OR
       p_receipt.effect_key IS DISTINCT FROM v_job.effect_idempotency_key OR
       p_receipt.provider IS DISTINCT FROM 'documents.adoption' OR
       p_receipt.operation IS DISTINCT FROM 'copy' OR NOT EXISTS (
           SELECT 1 FROM platform.job_attempts AS attempt
           WHERE attempt.job_id=v_job.id AND attempt.finished_at IS NULL
             AND attempt.effect_id=p_receipt.effect_id
             AND attempt.effect_request_sha256=p_receipt.request_sha256
       ) THEN
        RAISE EXCEPTION USING ERRCODE='P5603',
            MESSAGE='cutover effect receipt differs';
    END IF;
END;
$function$;`

func prepareNonMaintenanceCutoverGuard(t *testing.T,
	fixture analyticsDashboardExportPG18,
) {
	t.Helper()
	if _, err := fixture.database.ExecContext(
		fixture.ctx, nonMaintenanceCutoverReceiptGuardPG18,
	); err != nil {
		t.Fatalf("prepare cutover receipt guard: %v", err)
	}
}

func createNonMaintenanceExternalActive(t *testing.T,
	fixture analyticsDashboardExportPG18, kind, provider, operation string,
) analyticsMaintenanceActive {
	t.Helper()
	evidence := analyticsMaintenanceActive{
		kind: kind, jobID: analyticsExportV7(t), leaseID: analyticsExportV7(t),
		exportID: analyticsExportV7(t), effectID: uuid.New(),
	}
	payloadRef := evidence.exportID.String()
	_, err := fixture.runtime.Exec(fixture.ctx, `INSERT INTO platform.jobs(
    id,tenant_id,scope_kind,scope_id,authority_group_id,
    authority_group_generation,authority_domain,cutover_generation,
    writer_origin,kind,payload_revision,payload_ref,payload_sha256,dedupe_key,
    effect_class,effect_idempotency_key,priority,run_at,state,attempts,
    max_attempts,lease_generation
) VALUES ($1,$2,'tenant',$2,'analytics',1,'analytics',1,'v2',$3,1,$4,
    encode(sha256(convert_to($4,'UTF8')),'hex'),$3||':'||$4,
    'external',$3||':'||$4,10,$5,'queued',0,3,0)`, evidence.jobID,
		asyncTenantA, kind, payloadRef, time.Now().UTC().Add(-time.Second))
	if err != nil {
		t.Fatalf("seed non-maintenance job: %v", err)
	}
	identity := &workerIdentity{marker: 1}
	authority, err := newJobAuthority(identity, "tenant", uuid.MustParse(asyncTenantA),
		"analytics", 1, "analytics", 1)
	if err != nil {
		t.Fatal(err)
	}
	policy, err := jobsv1.NewKindPolicy(kind, 1, "analytics", jobsv1.EffectExternal, 3)
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := jobsv1.NewCatalog([]jobsv1.KindPolicy{policy})
	if err != nil {
		t.Fatal(err)
	}
	workerRepo := &workerRepository{
		database: fixture.runtime, cleanup: 5 * time.Second, identity: identity,
	}
	claims, err := workerRepo.ClaimJobs(
		fixture.ctx, authority, evidence.leaseID, 1, catalog,
	)
	if err != nil || len(claims) != 1 {
		t.Fatalf("claim non-maintenance job: claims=%d err=%v", len(claims), err)
	}
	evidence.claim, evidence.attemptID = claims[0], claims[0].AttemptID()
	transaction, err := fixture.runtime.BeginTx(fixture.ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer transaction.Rollback(fixture.ctx)
	bound := &repository{
		database:  lifecycleBoundPGTx{transaction: transaction},
		authority: lifecycleAuthorityStamp(authority),
	}
	spec, err := NewExternalEffectRequestSpec(evidence.effectID,
		provider, operation, evidence.jobID.String(), evidence.jobID)
	if err != nil {
		t.Fatal(err)
	}
	evidence.permit, err = bound.MarkExternalEffectStarted(
		fixture.ctx, evidence.claim, spec,
	)
	if err != nil {
		t.Fatalf("start non-maintenance effect: %v", err)
	}
	if err = transaction.Commit(fixture.ctx); err != nil {
		t.Fatal(err)
	}
	evidence.repository, evidence.authority = workerRepo, authority
	return evidence
}

func nonMaintenanceSucceededReceipt(t *testing.T,
	evidence analyticsMaintenanceActive,
) effectsv1.Receipt {
	t.Helper()
	request := evidence.permit.Request()
	providerRef := "provider/result/" + evidence.jobID.String()
	resultRef := "result/non-maintenance/" + evidence.jobID.String()
	digest := sha256.Sum256([]byte(resultRef))
	receipt, err := effectsv1.NewSucceededReceipt(request, uuid.New().String(),
		"DELIVERED", &providerRef, resultRef, hex.EncodeToString(digest[:]),
		time.Now().UTC().Format(time.RFC3339Nano))
	if err != nil {
		t.Fatal(err)
	}
	return receipt
}

func bindNonMaintenanceCutover(t *testing.T,
	fixture analyticsDashboardExportPG18, jobID uuid.UUID,
) {
	t.Helper()
	if _, err := fixture.database.ExecContext(fixture.ctx,
		"INSERT INTO platform.cutover_job_bindings(job_id) VALUES ($1)", jobID); err != nil {
		t.Fatal(err)
	}
}

func assertNonMaintenanceReceiptSucceeded(t *testing.T,
	fixture analyticsDashboardExportPG18, evidence analyticsMaintenanceActive,
	receipt effectsv1.Receipt,
) {
	t.Helper()
	var exact bool
	err := fixture.database.QueryRowContext(fixture.ctx, `SELECT
    job.state='succeeded' AND attempt.outcome='succeeded'
    AND attempt.effect_receipt_id=receipt.receipt_id
    AND receipt.provider=$3 AND receipt.operation=$4
FROM platform.jobs job JOIN platform.job_attempts attempt ON attempt.job_id=job.id
JOIN platform.idempotency_receipts receipt
  ON receipt.receipt_id=attempt.effect_receipt_id
WHERE job.id=$1 AND receipt.receipt_id=$2`, evidence.jobID, receipt.ReceiptID,
		receipt.Provider, receipt.Operation).Scan(&exact)
	if err != nil || !exact {
		t.Fatalf("non-maintenance receipt exact=%v err=%v", exact, err)
	}
}

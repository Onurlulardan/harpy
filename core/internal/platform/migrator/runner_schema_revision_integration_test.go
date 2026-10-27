//go:build integration

package migrator

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
)

const revisionEqualJobID = "018f0000-0000-7000-8000-000000000771"
const revisionZeroJobID = "018f0000-0000-7000-8000-000000000772"
const revisionTwoJobID = "018f0000-0000-7000-8000-000000000773"
const revisionRaceLeftJobID = "018f0000-0000-7000-8000-000000000774"
const revisionRaceRightJobID = "018f0000-0000-7000-8000-000000000775"

func assertSchemaRevisionMonotonicity(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	if err := insertQueuedRevisionJob(ctx, database, revisionEqualJobID,
		permanentFailureJobID, 1); err == nil {
		t.Fatal("equal terminal schema revision was accepted")
	}
	if err := insertQueuedRevisionJob(ctx, database, revisionZeroJobID,
		permanentFailureJobID, 0); err == nil {
		t.Fatal("lower schema revision was accepted")
	}
	if err := insertQueuedRevisionJob(ctx, database, revisionTwoJobID,
		permanentFailureJobID, 2); err != nil {
		t.Fatalf("next schema revision rejected: %v", err)
	}
}

func assertConcurrentSchemaRevision(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	start := make(chan struct{})
	results := make(chan error, 2)
	for _, jobID := range []string{revisionRaceLeftJobID, revisionRaceRightJobID} {
		go func(jobID string) {
			<-start
			results <- insertQueuedRevisionJob(
				ctx, database, jobID, activeAttemptEightJobID, 2,
			)
		}(jobID)
	}
	close(start)
	accepted := 0
	for range 2 {
		if err := <-results; err == nil {
			accepted++
		}
	}
	if accepted != 1 {
		t.Fatalf("concurrent next schema revision accepts = %d, want 1", accepted)
	}
}

func insertQueuedRevisionJob(
	ctx context.Context,
	database *sql.DB,
	jobID string,
	resourceID string,
	revision int64,
) error {
	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	planHash := integrationDigest(jobID + "-revision-plan")
	stepHash := integrationDigest(jobID + "-revision-step")
	_, err = tx.ExecContext(ctx, `INSERT INTO platform.schema_jobs
(job_id,tenant_id,owner,resource_kind,resource_id,resource_revision,
compiler_version,plan_contract_version,canonical_plan,plan_sha256,
step_count,available_at) VALUES($1,$2,'forms','form_storage',$3,$4,
'1.0.0','forms-schema-plan.v1','{"fixture":true}',$5,1,
pg_catalog.clock_timestamp())`, jobID, schemaTriggerTenantID,
		resourceID, revision, planHash)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO platform.schema_job_steps
(job_id,ordinal,plan_sha256,operation_id,transaction_class,executor_kind,
identifier_mapping,arguments,step_sha256,next_executor)
VALUES($1,0,$2,'verify_storage','transactional','migrator',
'{"fixture":true}','{}',$3,'migrator')`, jobID, planHash, stepHash)
	if err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit schema revision: %w", err)
	}
	return nil
}

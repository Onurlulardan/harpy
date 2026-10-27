//go:build integration

package migrator

import (
	"context"
	"database/sql"
	"testing"
)

const schemaTriggerTenantID = "018f0000-0000-7000-8000-000000000701"
const progressJobID = "018f0000-0000-7000-8000-000000000710"
const multiStepJobID = "018f0000-0000-7000-8000-000000000720"
const expiredLeasedJobID = "018f0000-0000-7000-8000-000000000730"
const expiredRunningJobID = "018f0000-0000-7000-8000-000000000740"
const permanentFailureJobID = "018f0000-0000-7000-8000-000000000750"
const activeAttemptEightJobID = "018f0000-0000-7000-8000-000000000760"

type schemaSeed struct {
	jobID, planHash, operation, stepHash string
	stepCount, currentOrdinal            int
}

func seedSchemaTriggerRows(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	setSchemaFixtureTriggers(t, ctx, database, false)
	defer setSchemaFixtureTriggers(t, context.Background(), database, true)
	_, err := database.ExecContext(ctx, `INSERT INTO platform.tenants
(id,name,slug,status,settings,owner_id,schema_name,created_at,updated_at)
VALUES($1,'Trigger tenant','trigger-tenant','ACTIVE','{}',
'018f0000-0000-7000-8000-000000000702','tenant_trigger',
pg_catalog.clock_timestamp(),pg_catalog.clock_timestamp())`, schemaTriggerTenantID)
	if err != nil {
		t.Fatal(err)
	}
	seeds := []schemaSeed{
		{progressJobID, integrationDigest("progress-plan"),
			"backfill_field_batch", integrationDigest("progress-step"), 1, 0},
		{multiStepJobID, integrationDigest("multi-plan"),
			"verify_storage", integrationDigest("multi-current"), 2, 1},
		{expiredLeasedJobID, integrationDigest("expired-leased-plan"),
			"verify_storage", integrationDigest("expired-leased-step"), 1, 0},
		{expiredRunningJobID, integrationDigest("expired-running-plan"),
			"verify_storage", integrationDigest("expired-running-step"), 1, 0},
		{permanentFailureJobID, integrationDigest("permanent-failure-plan"),
			"verify_storage", integrationDigest("permanent-failure-step"), 1, 0},
		{activeAttemptEightJobID, integrationDigest("active-eight-plan"),
			"verify_storage", integrationDigest("active-eight-step"), 1, 0},
	}
	for _, seed := range seeds {
		insertSchemaSeed(t, ctx, database, seed)
	}
	seedSchemaStates(t, ctx, database)
	seedExpiredAttemptHistory(t, ctx, database, seeds[2])
	seedExpiredAttemptHistory(t, ctx, database, seeds[3])
	seedExpiredAttemptHistory(t, ctx, database, seeds[5])
	setSchemaFixtureTriggers(t, ctx, database, true)
}

func insertSchemaSeed(t *testing.T, ctx context.Context, database *sql.DB, seed schemaSeed) {
	t.Helper()
	_, err := database.ExecContext(ctx, `INSERT INTO platform.schema_jobs
(job_id,tenant_id,owner,resource_kind,resource_id,resource_revision,
compiler_version,plan_contract_version,canonical_plan,plan_sha256,
step_count,available_at) VALUES($1,$2,'forms','form_storage',$3,1,
'1.0.0','forms-schema-plan.v1','{"fixture":true}',$4,$5,
pg_catalog.clock_timestamp())`, seed.jobID, schemaTriggerTenantID,
		seed.jobID, seed.planHash, seed.stepCount)
	if err != nil {
		t.Fatal(err)
	}
	for ordinal := 0; ordinal < seed.stepCount; ordinal++ {
		operation, stepHash := seed.operation, seed.stepHash
		if ordinal < seed.currentOrdinal {
			operation, stepHash = "verify_storage", integrationDigest(seed.jobID+"-previous")
		}
		_, err = database.ExecContext(ctx, `INSERT INTO platform.schema_job_steps
(job_id,ordinal,plan_sha256,operation_id,transaction_class,executor_kind,
identifier_mapping,arguments,step_sha256,next_executor)
VALUES($1,$2,$3,$4,'transactional','migrator','{"fixture":true}','{}',$5,'migrator')`,
			seed.jobID, ordinal, seed.planHash, operation, stepHash)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func setSchemaFixtureTriggers(
	t *testing.T, ctx context.Context, database *sql.DB, enabled bool,
) {
	t.Helper()
	action := "DISABLE"
	if enabled {
		action = "ENABLE"
	}
	statements := []string{
		"ALTER TABLE platform.schema_jobs " + action + " TRIGGER schema_jobs_validate_plan",
		"ALTER TABLE platform.schema_jobs " + action + " TRIGGER schema_jobs_validate_alignment",
		"ALTER TABLE platform.schema_jobs " + action + " TRIGGER schema_jobs_validate_transition_v2",
		"ALTER TABLE platform.schema_job_steps " + action + " TRIGGER schema_job_steps_validate_plan",
		"ALTER TABLE platform.schema_job_steps " + action + " TRIGGER schema_job_steps_validate_alignment",
		"ALTER TABLE platform.schema_job_steps " + action + " TRIGGER schema_job_steps_validate_transition_v3",
		"ALTER TABLE platform.schema_job_step_attempts " + action + " TRIGGER schema_job_step_attempts_validate_insert_v2",
	}
	for _, statement := range statements {
		if _, err := database.ExecContext(ctx, statement); err != nil {
			t.Fatalf("%s schema fixture trigger: %v", action, err)
		}
	}
}

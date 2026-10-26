//go:build integration

package maintenance

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"testing"

	maintenanceapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/maintenance"
	"github.com/google/uuid"
)

func assertMaintenancePG18ProfileExecution(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	fixtures := createMaintenancePG18ProfilePlans(t, ctx, database)
	assertMaintenancePG18ProfileClaimIsolation(t, ctx, database, fixtures)
	for index, fixture := range fixtures {
		assertMaintenancePG18ProfileSourceFence(t, ctx, database, fixture)
		assertMaintenancePG18ProfileActualSchemaCAS(
			t, ctx, database, fixture, index,
		)
		assertMaintenancePG18ProfileLivePreconversion(
			t, ctx, database, fixture, index,
		)
		pageOne := profileMaintenancePG18Claim(t, fixture, 0,
			fmt.Sprintf("72000000-0000-4000-8000-0000000000%d1", index))
		rolled, err := executeMaintenancePG18Command(t, ctx, database,
			fixture.tenant.schema, pageOne,
			uuid.MustParse("019c6a7e-7c00-7000-8000-000000000410"), false)
		if err != nil || rolled.Successor == nil {
			t.Fatalf("%s profile rollback=%#v/%v", fixture.tenant.schema, rolled, err)
		}
		assertMaintenancePG18ProfileRollback(t, ctx, database, fixture)
		first, err := executeMaintenancePG18Command(t, ctx, database,
			fixture.tenant.schema, pageOne,
			uuid.MustParse("019c6a7e-7c00-7000-8000-000000000411"), true)
		if err != nil || first.Successor == nil || first.Successor.JobID != fixture.jobs[1] {
			t.Fatalf("%s profile page one=%#v/%v", fixture.tenant.schema, first, err)
		}
		replay, err := executeMaintenancePG18Command(t, ctx, database,
			fixture.tenant.schema, pageOne,
			uuid.MustParse("019c6a7e-7c00-7000-8000-000000000412"), true)
		if err != nil || replay.Successor == nil || replay.Successor.JobID != fixture.jobs[1] {
			t.Fatalf("%s profile replay=%#v/%v", fixture.tenant.schema, replay, err)
		}
		assertMaintenancePG18ProfileLiveRaces(
			t, ctx, database, fixture, index,
		)
		pageTwo := profileMaintenancePG18Claim(t, fixture, 1,
			fmt.Sprintf("72000000-0000-4000-8000-0000000000%d2", index))
		terminal, err := executeMaintenancePG18Command(t, ctx, database,
			fixture.tenant.schema, pageTwo,
			uuid.MustParse("019c6a7e-7c00-7000-8000-000000000413"), true)
		if err != nil || terminal.Successor != nil {
			t.Fatalf("%s profile resume=%#v/%v", fixture.tenant.schema, terminal, err)
		}
		assertMaintenancePG18ProfileParity(t, ctx, database, fixture, index)
	}
	assertMaintenancePG18ProfileInvalidRowFence(t, ctx, database, fixtures[0])
}

func profileMaintenancePG18Claim(
	t *testing.T, fixture maintenancePlanFixture, page int, attempt string,
) maintenanceapp.Claim {
	t.Helper()
	reference, digest, err := maintenanceapp.ProfileReference(
		uuid.MustParse(fixture.tenant.id), fixture.planID, int64(page+1),
	)
	if err != nil {
		t.Fatal(err)
	}
	return maintenanceapp.Claim{
		JobID: fixture.jobs[page], AttemptID: uuid.MustParse(attempt),
		TenantID:         uuid.MustParse(fixture.tenant.id),
		Kind:             maintenanceapp.KindProfileBackfill,
		PayloadRevision:  maintenanceapp.PayloadRevision,
		PayloadReference: reference, PayloadSHA256: digest,
	}
}

func assertMaintenancePG18ProfileRollback(
	t *testing.T, ctx context.Context, database *sql.DB,
	fixture maintenancePlanFixture,
) {
	t.Helper()
	var schemaID uuid.UUID
	var revision int64
	var pageState string
	var completed, receipts, preconversions, capabilities int
	err := database.QueryRowContext(ctx, `SELECT usr.profile_schema_version_id,
 usr.revision,page.state,checkpoint.completed_pages,
 (SELECT pg_catalog.count(*) FROM `+fixture.tenant.schema+`.iam_maintenance_receipts
  WHERE job_id=$2),
 (SELECT pg_catalog.count(*) FROM `+fixture.tenant.schema+`.iam_profile_backfill_preconversions
  WHERE plan_id=$1 AND resource_id=$3),
 (SELECT pg_catalog.count(*) FROM `+fixture.tenant.schema+`.iam_profile_backfill_worker_capabilities)
FROM `+fixture.tenant.schema+`.iam_users AS usr
JOIN `+fixture.tenant.schema+`.iam_profile_backfill_pages AS page
 ON page.plan_id=$1 AND page.page_number=1
JOIN `+fixture.tenant.schema+`.iam_profile_backfill_checkpoints AS checkpoint
 ON checkpoint.plan_id=page.plan_id WHERE usr.id=$3`,
		fixture.planID, fixture.jobs[0], maintenanceUserOne).Scan(
		&schemaID, &revision, &pageState, &completed, &receipts,
		&preconversions, &capabilities)
	if err != nil || schemaID != fixture.targetSchema.ID() || revision != 4 ||
		pageState != "queued" || completed != 0 || receipts != 0 ||
		preconversions != 1 || capabilities != 0 {
		t.Fatalf("%s profile rollback=%s/%d/%s/%d/%d/%d/%d: %v",
			fixture.tenant.schema, schemaID, revision, pageState, completed,
			receipts, preconversions, capabilities, err)
	}
}

func assertMaintenancePG18ProfileParity(
	t *testing.T, ctx context.Context, database *sql.DB,
	fixture maintenancePlanFixture, tenantIndex int,
) {
	t.Helper()
	var planState, sourceState, targetState string
	var pages, converted, failures int
	err := database.QueryRowContext(ctx, `SELECT plan.state,source.state,target.state,
 checkpoint.completed_pages,checkpoint.converted_count,checkpoint.error_count
FROM `+fixture.tenant.schema+`.iam_profile_backfill_plans AS plan
JOIN `+fixture.tenant.schema+`.iam_profile_schema_versions AS source
 ON source.id=plan.source_schema_id
JOIN `+fixture.tenant.schema+`.iam_profile_schema_versions AS target
 ON target.id=plan.target_schema_id
JOIN `+fixture.tenant.schema+`.iam_profile_backfill_checkpoints AS checkpoint
 ON checkpoint.plan_id=plan.id WHERE plan.id=$1`, fixture.planID).Scan(
		&planState, &sourceState, &targetState, &pages, &converted, &failures)
	if err != nil || planState != "activated" || sourceState != "retired" ||
		targetState != "active" || pages != 2 || converted != 3 || failures != 0 {
		t.Fatalf("%s profile parity=%s/%s/%s/%d/%d/%d: %v",
			fixture.tenant.schema, planState, sourceState, targetState,
			pages, converted, failures, err)
	}
	for index, userID := range maintenanceProfileUsers {
		_, digest := maintenanceMixedTargetDocument(tenantIndex, index)
		var schemaID uuid.UUID
		var actualSHA []byte
		var revision int64
		err := database.QueryRowContext(ctx, `SELECT profile_schema_version_id,
profile_sha256,revision FROM `+fixture.tenant.schema+`.iam_users WHERE id=$1`,
			userID).Scan(&schemaID, &actualSHA, &revision)
		wantRevision := int64(4)
		if index == 1 {
			wantRevision = 5
		}
		if err != nil || schemaID != fixture.targetSchema.ID() ||
			revision != wantRevision ||
			!bytes.Equal(actualSHA, digest[:]) {
			t.Fatalf("%s user %s profile parity=%s/%d/%v",
				fixture.tenant.schema, userID, schemaID, revision, err)
		}
	}
	var retiredState string
	err = database.QueryRowContext(ctx, `SELECT state FROM `+
		fixture.tenant.schema+`.iam_profile_fields WHERE id=$1`,
		maintenanceProfileField).Scan(&retiredState)
	if err != nil || retiredState != "retired" {
		t.Fatalf("%s terminal field retirement=%q/%v",
			fixture.tenant.schema, retiredState, err)
	}
}

func assertMaintenancePG18ProfileSourceFence(
	t *testing.T, ctx context.Context, database *sql.DB,
	fixture maintenancePlanFixture,
) {
	t.Helper()
	_, err := database.ExecContext(ctx, `UPDATE `+fixture.tenant.schema+`.iam_users
SET profile_values='{}'::jsonb,
 profile_sha256=pg_catalog.sha256(pg_catalog.convert_to('{}','UTF8')),
 revision=revision+1 WHERE id=$1`, maintenanceUserTwo)
	if err == nil {
		t.Fatalf("%s live profile source write passed", fixture.tenant.schema)
	}
}

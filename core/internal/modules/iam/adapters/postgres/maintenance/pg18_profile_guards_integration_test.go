//go:build integration

package maintenance

import (
	"bytes"
	"context"
	"database/sql"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func assertMaintenancePG18ProfileClaimIsolation(
	t *testing.T, ctx context.Context, database *sql.DB,
	fixtures []maintenancePlanFixture,
) {
	t.Helper()
	foreign := profileMaintenancePG18Claim(t, fixtures[1], 0,
		"72000000-0000-4000-8000-000000000081")
	if _, err := executeMaintenancePG18Command(t, ctx, database,
		fixtures[0].tenant.schema, foreign,
		uuid.MustParse("019c6a7e-7c00-7000-8000-000000000481"), true,
	); err == nil {
		t.Fatal("foreign tenant profile claim passed")
	}
}

func assertMaintenancePG18ProfileActualSchemaCAS(
	t *testing.T, ctx context.Context, database *sql.DB,
	fixture maintenancePlanFixture, tenantIndex int,
) {
	t.Helper()
	claim := profileMaintenancePG18Claim(t, fixture, 0,
		"72000000-0000-4000-8000-000000000082")
	claimArguments, err := claimValues(claim)
	if err != nil {
		t.Fatal(err)
	}
	_, sourceSHA := maintenanceMixedSourceDocument(tenantIndex, 0)
	targetDocument, targetSHA := maintenanceMixedTargetDocument(tenantIndex, 0)
	transaction, err := database.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer transaction.Rollback()
	for _, statement := range []string{
		"SET LOCAL SESSION AUTHORIZATION harpy_v2_runtime",
		"SET LOCAL ROLE harpy_runtime",
	} {
		if _, err = transaction.ExecContext(ctx, statement); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = transaction.ExecContext(ctx,
		"SELECT pg_catalog.set_config('search_path',$1,true)",
		fixture.tenant.schema+", platform, pg_catalog"); err != nil {
		t.Fatal(err)
	}
	var revision int64
	err = transaction.QueryRowContext(ctx, `SELECT
iam_maintenance_profile_backfill_apply_v3(
 pg_catalog.current_schema(),$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		claimArguments[0], claimArguments[1], claimArguments[2], claimArguments[3],
		claimArguments[4], claimArguments[5], fixture.planID, 1,
		uuid.MustParse(maintenanceUserOne),
		uuid.MustParse(maintenanceOptionalSchemaID), int64(3), sourceSHA[:],
		[]byte(targetDocument), targetSHA[:],
	).Scan(&revision)
	if err == nil || !strings.Contains(strings.ToLower(err.Error()), "source changed") {
		t.Fatalf("%s profile actual-schema CAS error=%v", fixture.tenant.schema, err)
	}
}

func assertMaintenancePG18ProfileInvalidRowFence(
	t *testing.T, ctx context.Context, database *sql.DB,
	source maintenancePlanFixture,
) {
	t.Helper()
	fixture := createMaintenancePG18InvalidProfilePlan(t, ctx, database, source)
	pageOne := profileMaintenancePG18Claim(t, fixture, 0,
		"72000000-0000-4000-8000-000000000083")
	first, err := executeMaintenancePG18Command(t, ctx, database,
		fixture.tenant.schema, pageOne,
		uuid.MustParse("019c6a7e-7c00-7000-8000-000000000483"), true)
	if err != nil || first.Successor == nil || first.Successor.JobID != fixture.jobs[1] {
		t.Fatalf("profile compatibility page outcome=%#v/%v", first, err)
	}
	pageTwo := profileMaintenancePG18Claim(t, fixture, 1,
		"72000000-0000-4000-8000-000000000084")
	terminal, err := executeMaintenancePG18Command(t, ctx, database,
		fixture.tenant.schema, pageTwo,
		uuid.MustParse("019c6a7e-7c00-7000-8000-000000000484"), true)
	if err != nil || terminal.Successor != nil {
		t.Fatalf("invalid profile row outcome=%#v/%v", terminal, err)
	}
	_, sourceSHA := maintenanceMixedTargetDocument(0, 2)
	wantEvidence := maintenancePageDigest([][]byte{
		[]byte(maintenanceUserThree), []byte(source.targetSchema.ID().String()),
		sourceSHA[:], []byte("document_invalid"),
	})
	var planState, pageState, sourceState, targetState, errorCode string
	var pages, converted, errors int
	var evidence []byte
	var userSchema uuid.UUID
	var userRevision int64
	err = database.QueryRowContext(ctx, `SELECT plan.state,page.state,
 source_schema.state,target_schema.state,checkpoint.completed_pages,
 checkpoint.converted_count,checkpoint.error_count,failure.error_code,
 failure.evidence_sha256,usr.profile_schema_version_id,usr.revision
FROM `+fixture.tenant.schema+`.iam_profile_backfill_plans AS plan
JOIN `+fixture.tenant.schema+`.iam_profile_backfill_pages AS page
 ON page.plan_id=plan.id AND page.page_number=2
JOIN `+fixture.tenant.schema+`.iam_profile_backfill_checkpoints AS checkpoint
 ON checkpoint.plan_id=plan.id
JOIN `+fixture.tenant.schema+`.iam_profile_backfill_errors AS failure
 ON failure.plan_id=plan.id AND failure.page_number=page.page_number
JOIN `+fixture.tenant.schema+`.iam_profile_schema_versions AS source_schema
 ON source_schema.id=plan.source_schema_id
JOIN `+fixture.tenant.schema+`.iam_profile_schema_versions AS target_schema
 ON target_schema.id=plan.target_schema_id
JOIN `+fixture.tenant.schema+`.iam_users AS usr ON usr.id=failure.resource_id
WHERE plan.id=$1`, fixture.planID).Scan(
		&planState, &pageState, &sourceState, &targetState, &pages,
		&converted, &errors, &errorCode, &evidence, &userSchema, &userRevision)
	if err != nil || planState != "fenced" || pageState != "fenced" ||
		sourceState != "active" || targetState != "draft" || pages != 2 ||
		converted != 3 || errors != 1 || errorCode != "document_invalid" ||
		!bytes.Equal(evidence, wantEvidence[:]) || userSchema != source.targetSchema.ID() ||
		userRevision != 4 {
		t.Fatalf("invalid profile fence=%s/%s/%s/%s/%d/%d/%d/%s/%v/%s/%d: %v",
			planState, pageState, sourceState, targetState, pages, converted,
			errors, errorCode, bytes.Equal(evidence, wantEvidence[:]),
			userSchema, userRevision, err)
	}
	_, convertedSHA := maintenanceInvalidTargetDocument(0, 0)
	var convertedSchema uuid.UUID
	var actualSHA []byte
	var convertedRevision int64
	err = database.QueryRowContext(ctx, `SELECT profile_schema_version_id,
profile_sha256,revision FROM `+fixture.tenant.schema+`.iam_users WHERE id=$1`,
		maintenanceUserOne).Scan(&convertedSchema, &actualSHA, &convertedRevision)
	if err != nil || convertedSchema != fixture.targetSchema.ID() ||
		convertedRevision != 5 || !bytes.Equal(actualSHA, convertedSHA[:]) {
		t.Fatalf("retired compatibility apply=%s/%d/%v/%v",
			convertedSchema, convertedRevision,
			bytes.Equal(actualSHA, convertedSHA[:]), err)
	}
}

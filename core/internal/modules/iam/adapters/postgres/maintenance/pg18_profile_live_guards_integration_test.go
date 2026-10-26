//go:build integration

package maintenance

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"testing"

	"github.com/google/uuid"
)

func assertMaintenancePG18ProfileDirectTransitionDenied(
	t *testing.T, ctx context.Context, database *sql.DB,
	fixture maintenancePlanFixture, tenantIndex int,
) {
	t.Helper()
	document, digest := maintenanceMixedTargetDocument(tenantIndex, 2)
	transaction, err := database.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, statement := range []string{
		"GRANT SELECT, UPDATE ON iam_users TO harpy_runtime",
		"SET LOCAL SESSION AUTHORIZATION harpy_v2_runtime",
		"SET LOCAL ROLE harpy_runtime",
	} {
		if _, err = transaction.ExecContext(ctx,
			"SELECT pg_catalog.set_config('search_path',$1,true)",
			fixture.tenant.schema+", platform, pg_catalog"); err != nil {
			_ = transaction.Rollback()
			t.Fatal(err)
		}
		if _, err = transaction.ExecContext(ctx, statement); err != nil {
			_ = transaction.Rollback()
			t.Fatal(err)
		}
	}
	_, err = transaction.ExecContext(ctx, `UPDATE iam_users
SET profile_schema_version_id=$2,profile_values=$3::jsonb,
 profile_sha256=$4,revision=revision+1
WHERE id=$1 AND profile_schema_version_id=$5 AND revision=3`,
		maintenanceUserThree, fixture.targetSchema.ID(), document,
		digest[:], maintenanceUserSchemaID,
	)
	_ = transaction.Rollback()
	requireProfileGuardError(t, err, "lacks authority evidence")
	var schemaID uuid.UUID
	var revision int64
	err = database.QueryRowContext(ctx, `SELECT profile_schema_version_id,revision
FROM `+fixture.tenant.schema+`.iam_users WHERE id=$1`,
		maintenanceUserThree).Scan(&schemaID, &revision)
	if err != nil || schemaID != uuid.MustParse(maintenanceUserSchemaID) ||
		revision != 3 {
		t.Fatalf("%s direct transition mutated source=%s/%d: %v",
			fixture.tenant.schema, schemaID, revision, err)
	}
}

func assertMaintenancePG18TargetPatchRollback(
	t *testing.T, ctx context.Context, database *sql.DB,
	fixture maintenancePlanFixture, originalSHA [32]byte,
) {
	t.Helper()
	edited := fmt.Sprintf(`{"%s":"Edited","%s":"Edited"}`,
		maintenanceProfileField, maintenanceTargetFieldID)
	editedSHA := sha256.Sum256([]byte(edited))
	operationID := uuid.MustParse(map[string]string{
		"t_iam_a": "5a000000-0000-4000-8000-000000000001",
		"t_iam_b": "5a000000-0000-4000-8000-000000000002",
	}[fixture.tenant.schema])
	requestSHA := sha256.Sum256([]byte("target-profile-patch-" +
		fixture.tenant.id))
	transaction := beginProfileRuntimeTransaction(
		t, ctx, database, fixture.tenant.schema,
	)
	var resultID uuid.UUID
	var revision, clock int64
	var replayed bool
	err := transaction.QueryRowContext(ctx, `SELECT * FROM
iam_identity_user_update_v2(
 pg_catalog.current_schema(),$1,$2,$3,4,$4,$5,$6,$7,$8,1,1)`,
		operationID, requestSHA[:], maintenanceUserOne,
		fixture.targetSchema.ID(), []byte(edited), editedSHA[:],
		maintenanceUserThree, maintenanceReadPermission,
	).Scan(&resultID, &revision, &clock, &replayed)
	_ = transaction.Rollback()
	if err != nil || resultID != uuid.MustParse(maintenanceUserOne) ||
		revision != 5 || clock < 1 || replayed {
		t.Fatalf("%s target patch rollback=%s/%d/%d/%t: %v",
			fixture.tenant.schema, resultID, revision, clock, replayed, err)
	}
	assertMaintenancePG18ProfileTableACL(t, ctx, database, fixture)
	var schemaID uuid.UUID
	var storedRevision int64
	var storedSHA []byte
	err = database.QueryRowContext(ctx, `SELECT profile_schema_version_id,
 revision,profile_sha256 FROM `+fixture.tenant.schema+`.iam_users
WHERE id=$1`, maintenanceUserOne).Scan(
		&schemaID, &storedRevision, &storedSHA,
	)
	if err != nil || schemaID != fixture.targetSchema.ID() || storedRevision != 4 ||
		!bytes.Equal(storedSHA, originalSHA[:]) {
		t.Fatalf("%s target rollback state=%s/%d/%x: %v",
			fixture.tenant.schema, schemaID, storedRevision, storedSHA, err)
	}
}

func assertMaintenancePG18ProfileTableACL(
	t *testing.T, ctx context.Context, database *sql.DB,
	fixture maintenancePlanFixture,
) {
	t.Helper()
	for _, table := range []string{
		"iam_profile_backfill_preconversions",
		"iam_profile_backfill_worker_capabilities",
	} {
		transaction := beginProfileRuntimeTransaction(
			t, ctx, database, fixture.tenant.schema,
		)
		_, err := transaction.ExecContext(ctx,
			"DELETE FROM "+table+" WHERE false")
		_ = transaction.Rollback()
		requireProfileGuardError(t, err, "permission denied")
	}
}

func assertMaintenancePG18CapabilityLeftoverDenied(
	t *testing.T, ctx context.Context, database *sql.DB,
	fixture maintenancePlanFixture, tenantIndex int,
) {
	t.Helper()
	_, sourceSHA := maintenanceMixedSourceDocument(tenantIndex, 1)
	_, targetSHA := maintenanceMixedTargetDocument(tenantIndex, 1)
	transaction, err := database.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = transaction.ExecContext(ctx,
		"SELECT pg_catalog.set_config('search_path',$1,true)",
		fixture.tenant.schema+", pg_catalog")
	if err == nil {
		_, err = transaction.ExecContext(ctx, `INSERT INTO
iam_profile_backfill_worker_capabilities(
 backend_pid,transaction_id,plan_id,page_number,resource_kind,resource_id,
 actual_source_schema_id,source_revision,source_sha256,target_schema_id,
 target_sha256,created_at)
VALUES(pg_catalog.pg_backend_pid(),pg_catalog.pg_current_xact_id(),$1,2,
 'user',$2,$3,4,$4,$5,$6,pg_catalog.transaction_timestamp())`,
			fixture.planID, maintenanceUserTwo, maintenanceOptionalSchemaID,
			sourceSHA[:], fixture.targetSchema.ID(), targetSHA[:])
	}
	if err == nil {
		err = transaction.Commit()
	} else {
		_ = transaction.Rollback()
	}
	requireProfileGuardError(t, err, "was not consumed")
	var count int
	err = database.QueryRowContext(ctx, `SELECT pg_catalog.count(*) FROM `+
		fixture.tenant.schema+`.iam_profile_backfill_worker_capabilities`).
		Scan(&count)
	if err != nil || count != 0 {
		t.Fatalf("%s leftover capability count=%d: %v",
			fixture.tenant.schema, count, err)
	}
}

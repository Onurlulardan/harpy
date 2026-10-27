//go:build integration

package maintenance

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func assertMaintenancePG18ProfileLivePreconversion(
	t *testing.T, ctx context.Context, database *sql.DB,
	fixture maintenancePlanFixture, tenantIndex int,
) {
	t.Helper()
	assertMaintenancePG18ProfileNewUser(
		t, ctx, database, fixture, tenantIndex,
	)
	_, sourceSHA := maintenanceMixedSourceDocument(tenantIndex, 0)
	targetDocument, targetSHA := maintenanceMixedTargetDocument(tenantIndex, 0)
	operationID := uuid.MustParse([]string{
		"58000000-0000-4000-8000-000000000001",
		"58000000-0000-4000-8000-000000000002",
	}[tenantIndex])
	requestSHA := sha256.Sum256([]byte("live-profile-update-" + fixture.tenant.id))
	transaction := beginProfileRuntimeTransaction(
		t, ctx, database, fixture.tenant.schema,
	)
	var pageNumber int
	err := transaction.QueryRowContext(ctx, `SELECT
iam_profile_preconversion_reserve_v1(
 pg_catalog.current_schema(),$1,'user',$2,$3,$4,$5,$6,$7)`,
		fixture.planID, maintenanceUserOne, maintenanceUserSchemaID,
		int64(3), sourceSHA[:], targetSHA[:], targetSHA[:],
	).Scan(&pageNumber)
	if err == nil {
		var resultID uuid.UUID
		var revision, authorizationClock int64
		var replayed bool
		err = transaction.QueryRowContext(ctx, `SELECT * FROM
iam_identity_user_update_v2(
 pg_catalog.current_schema(),$1,$2,$3,3,$4,$5,$6,$7,$8,1,1)`,
			operationID, requestSHA[:], maintenanceUserOne,
			fixture.targetSchema.ID(), []byte(targetDocument), targetSHA[:],
			maintenanceUserThree, maintenanceReadPermission,
		).Scan(&resultID, &revision, &authorizationClock, &replayed)
		if err == nil && (resultID != uuid.MustParse(maintenanceUserOne) ||
			revision != 4 || authorizationClock < 1 || replayed) {
			err = ErrRuntimeInvalid
		}
	}
	if err == nil && pageNumber == 1 {
		err = transaction.Commit()
	} else {
		_ = transaction.Rollback()
	}
	if err != nil || pageNumber != 1 {
		t.Fatalf("%s live preconversion page=%d: %v",
			fixture.tenant.schema, pageNumber, err)
	}
	assertMaintenancePG18ProfileUpdateReplay(
		t, ctx, database, fixture, operationID, requestSHA,
		targetDocument, targetSHA,
	)
	assertMaintenancePG18PreconversionStored(
		t, ctx, database, fixture, sourceSHA, targetSHA,
	)
	assertMaintenancePG18ProfileDirectTransitionDenied(
		t, ctx, database, fixture, tenantIndex,
	)
	assertMaintenancePG18TargetPatchRollback(
		t, ctx, database, fixture, targetSHA,
	)
	assertMaintenancePG18CapabilityLeftoverDenied(
		t, ctx, database, fixture, tenantIndex,
	)
}

func assertMaintenancePG18ProfileUpdateReplay(
	t *testing.T, ctx context.Context, database *sql.DB,
	fixture maintenancePlanFixture, operationID uuid.UUID,
	requestSHA [32]byte, document string, digest [32]byte,
) {
	t.Helper()
	transaction := beginProfileRuntimeTransaction(
		t, ctx, database, fixture.tenant.schema,
	)
	var resultID uuid.UUID
	var revision, clock int64
	var replayed bool
	err := transaction.QueryRowContext(ctx, `SELECT * FROM
iam_identity_user_update_v2(
 pg_catalog.current_schema(),$1,$2,$3,3,$4,$5,$6,$7,$8,1,1)`,
		operationID, requestSHA[:], maintenanceUserOne,
		fixture.targetSchema.ID(), []byte(document), digest[:],
		maintenanceUserThree, maintenanceReadPermission,
	).Scan(&resultID, &revision, &clock, &replayed)
	if err == nil {
		err = transaction.Commit()
	} else {
		_ = transaction.Rollback()
	}
	if err != nil || resultID != uuid.MustParse(maintenanceUserOne) ||
		revision != 4 || clock < 1 || !replayed {
		t.Fatalf("%s profile update replay=%s/%d/%d/%t: %v",
			fixture.tenant.schema, resultID, revision, clock, replayed, err)
	}
}

func assertMaintenancePG18PreconversionStored(
	t *testing.T, ctx context.Context, database *sql.DB,
	fixture maintenancePlanFixture, sourceSHA, targetSHA [sha256.Size]byte,
) {
	t.Helper()
	var actualSchema, rowSchema uuid.UUID
	var sourceRevision, firstRevision, rowRevision int64
	var storedSource, planned, first, rowSHA []byte
	var capabilities int
	err := database.QueryRowContext(ctx, `SELECT
 receipt.actual_source_schema_id,receipt.source_revision,
 receipt.source_sha256,receipt.planned_target_sha256,
 receipt.first_target_revision,receipt.first_target_sha256,
 usr.profile_schema_version_id,usr.revision,usr.profile_sha256,
 (SELECT pg_catalog.count(*) FROM `+fixture.tenant.schema+
		`.iam_profile_backfill_worker_capabilities)
FROM `+fixture.tenant.schema+`.iam_profile_backfill_preconversions AS receipt
JOIN `+fixture.tenant.schema+`.iam_users AS usr
 ON usr.id=receipt.resource_id
WHERE receipt.plan_id=$1 AND receipt.resource_id=$2`,
		fixture.planID, maintenanceUserOne).Scan(
		&actualSchema, &sourceRevision, &storedSource, &planned,
		&firstRevision, &first, &rowSchema, &rowRevision, &rowSHA,
		&capabilities,
	)
	if err != nil || actualSchema != uuid.MustParse(maintenanceUserSchemaID) ||
		sourceRevision != 3 || firstRevision != 4 ||
		rowSchema != fixture.targetSchema.ID() || rowRevision != 4 ||
		!bytes.Equal(storedSource, sourceSHA[:]) ||
		!bytes.Equal(planned, targetSHA[:]) ||
		!bytes.Equal(first, targetSHA[:]) ||
		!bytes.Equal(rowSHA, targetSHA[:]) || capabilities != 0 {
		t.Fatalf("%s preconversion evidence invalid: %v",
			fixture.tenant.schema, err)
	}
}

func requireProfileGuardError(t *testing.T, err error, fragment string) {
	t.Helper()
	if err == nil || !strings.Contains(strings.ToLower(err.Error()), fragment) {
		t.Fatalf("profile guard error=%v, want %q", err, fragment)
	}
}

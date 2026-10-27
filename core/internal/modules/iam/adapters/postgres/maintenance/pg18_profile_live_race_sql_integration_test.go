//go:build integration

package maintenance

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"testing"

	maintenanceapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/maintenance"
	"github.com/google/uuid"
)

func executeMaintenancePG18RaceWorker(
	ctx context.Context, transaction *sql.Tx, fixture maintenancePlanFixture,
	tenantIndex, userIndex int,
) (int64, error) {
	claim, err := profileMaintenancePG18RaceClaim(
		fixture, tenantIndex, userIndex,
	)
	if err != nil {
		return 0, err
	}
	claimArguments, err := claimValues(claim)
	if err != nil {
		return 0, err
	}
	sourceSchema, sourceRevision := maintenancePG18RaceSource(userIndex)
	_, sourceSHA := maintenanceMixedSourceDocument(tenantIndex, userIndex)
	targetDocument, targetSHA := maintenanceMixedTargetDocument(
		tenantIndex, userIndex,
	)
	var revision int64
	err = transaction.QueryRowContext(ctx, `SELECT
iam_maintenance_profile_backfill_apply_v3(
 pg_catalog.current_schema(),$1,$2,$3,$4,$5,$6,$7,2,$8,$9,$10,$11,$12,$13)`,
		claimArguments[0], claimArguments[1], claimArguments[2],
		claimArguments[3], claimArguments[4], claimArguments[5],
		fixture.planID, uuid.MustParse(maintenanceProfileUsers[userIndex]),
		uuid.MustParse(sourceSchema), sourceRevision, sourceSHA[:],
		[]byte(targetDocument), targetSHA[:],
	).Scan(&revision)
	return revision, err
}

func executeMaintenancePG18RaceLiveWrite(
	ctx context.Context, transaction *sql.Tx, fixture maintenancePlanFixture,
	tenantIndex, userIndex int,
) (int, int64, error) {
	sourceSchema, sourceRevision := maintenancePG18RaceSource(userIndex)
	_, sourceSHA := maintenanceMixedSourceDocument(tenantIndex, userIndex)
	targetDocument, targetSHA := maintenanceMixedTargetDocument(
		tenantIndex, userIndex,
	)
	resourceID := uuid.MustParse(maintenanceProfileUsers[userIndex])
	var page int
	err := transaction.QueryRowContext(ctx, `SELECT
iam_profile_preconversion_reserve_v1(
 pg_catalog.current_schema(),$1,'user',$2,$3,$4,$5,$6,$6)`,
		fixture.planID, resourceID, uuid.MustParse(sourceSchema),
		sourceRevision, sourceSHA[:], targetSHA[:],
	).Scan(&page)
	if err != nil {
		return 0, 0, err
	}
	operationID := uuid.MustParse(fmt.Sprintf(
		"5b000000-0000-4000-8000-%012d",
		(tenantIndex+1)*10+userIndex,
	))
	requestSHA := sha256.Sum256([]byte(fmt.Sprintf(
		"profile-live-race-%s-%d", fixture.tenant.id, userIndex,
	)))
	var resultID uuid.UUID
	var revision, clock int64
	var replayed bool
	err = transaction.QueryRowContext(ctx, `SELECT * FROM
iam_identity_user_update_v2(
 pg_catalog.current_schema(),$1,$2,$3,$4,$5,$6,$7,$8,$9,1,1)`,
		operationID, requestSHA[:], resourceID, sourceRevision,
		fixture.targetSchema.ID(), []byte(targetDocument), targetSHA[:],
		uuid.MustParse(maintenanceUserThree),
		uuid.MustParse(maintenanceReadPermission),
	).Scan(&resultID, &revision, &clock, &replayed)
	if err == nil && (resultID != resourceID || revision != sourceRevision+1 ||
		clock < 1 || replayed) {
		err = ErrRuntimeInvalid
	}
	return page, revision, err
}

func profileMaintenancePG18RaceClaim(
	fixture maintenancePlanFixture, tenantIndex, userIndex int,
) (maintenanceapp.Claim, error) {
	tenantID := uuid.MustParse(fixture.tenant.id)
	reference, digest, err := maintenanceapp.ProfileReference(
		tenantID, fixture.planID, 2,
	)
	if err != nil {
		return maintenanceapp.Claim{}, err
	}
	return maintenanceapp.Claim{
		TenantID: tenantID, JobID: fixture.jobs[1],
		AttemptID: uuid.MustParse(fmt.Sprintf(
			"73000000-0000-4000-8000-%012d",
			(tenantIndex+1)*10+userIndex,
		)),
		Kind:             maintenanceapp.KindProfileBackfill,
		PayloadRevision:  maintenanceapp.PayloadRevision,
		PayloadReference: reference, PayloadSHA256: digest,
	}, nil
}

func maintenancePG18RaceSource(userIndex int) (string, int64) {
	if userIndex == 1 {
		return maintenanceOptionalSchemaID, 4
	}
	return maintenanceUserSchemaID, 3
}

func assertMaintenancePG18ProfileRaceStored(
	t *testing.T, ctx context.Context, database *sql.DB,
	fixture maintenancePlanFixture, tenantIndex int,
) {
	t.Helper()
	for _, userIndex := range []int{1, 2} {
		sourceSchema, sourceRevision := maintenancePG18RaceSource(userIndex)
		_, sourceSHA := maintenanceMixedSourceDocument(tenantIndex, userIndex)
		_, targetSHA := maintenanceMixedTargetDocument(tenantIndex, userIndex)
		var receiptSchema, rowSchema uuid.UUID
		var receiptRevision, firstRevision, rowRevision int64
		var storedSource, planned, first, current []byte
		err := database.QueryRowContext(ctx, `SELECT
receipt.actual_source_schema_id,receipt.source_revision,receipt.source_sha256,
receipt.planned_target_sha256,receipt.first_target_revision,
receipt.first_target_sha256,usr.profile_schema_version_id,usr.revision,
usr.profile_sha256 FROM `+fixture.tenant.schema+`.
iam_profile_backfill_preconversions AS receipt JOIN `+
			fixture.tenant.schema+`.iam_users AS usr ON usr.id=receipt.resource_id
WHERE receipt.plan_id=$1 AND receipt.resource_id=$2`,
			fixture.planID, maintenanceProfileUsers[userIndex]).Scan(
			&receiptSchema, &receiptRevision, &storedSource, &planned,
			&firstRevision, &first, &rowSchema, &rowRevision, &current,
		)
		if err != nil || receiptSchema != uuid.MustParse(sourceSchema) ||
			receiptRevision != sourceRevision || firstRevision != sourceRevision+1 ||
			rowSchema != fixture.targetSchema.ID() || rowRevision != sourceRevision+1 ||
			!bytes.Equal(storedSource, sourceSHA[:]) ||
			!bytes.Equal(planned, targetSHA[:]) || !bytes.Equal(first, targetSHA[:]) ||
			!bytes.Equal(current, targetSHA[:]) {
			t.Fatalf("%s race evidence user=%d: %v",
				fixture.tenant.schema, userIndex, err)
		}
	}
	var capabilities int
	if err := database.QueryRowContext(ctx, `SELECT pg_catalog.count(*) FROM `+
		fixture.tenant.schema+`.iam_profile_backfill_worker_capabilities`).
		Scan(&capabilities); err != nil || capabilities != 0 {
		t.Fatalf("%s race capabilities=%d: %v",
			fixture.tenant.schema, capabilities, err)
	}
}

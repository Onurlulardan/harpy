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

const (
	maintenanceLiveUser       = "52000000-0000-4000-8000-000000000001"
	maintenanceLiveIdentifier = "52000000-0000-4000-8000-000000000002"
	maintenanceLiveAccessMode = "52000000-0000-4000-8000-000000000003"
	maintenanceLiveAssignment = "52000000-0000-4000-8000-000000000004"
)

func assertMaintenancePG18ProfileNewUser(
	t *testing.T, ctx context.Context, database *sql.DB,
	fixture maintenancePlanFixture, tenantIndex int,
) {
	t.Helper()
	name := fmt.Sprintf("Live Tenant %d User", tenantIndex+1)
	document := fmt.Sprintf(`{"%s":%q,"%s":%q}`,
		maintenanceProfileField, name, maintenanceTargetFieldID, name)
	profileSHA := sha256.Sum256([]byte(document))
	email := fmt.Sprintf("live-%d@example.test", tenantIndex+1)
	emailSHA := sha256.Sum256([]byte(email))
	identifierHMAC := sha256.Sum256([]byte("live-identifier-" + fixture.tenant.id))
	requestSHA := sha256.Sum256([]byte("live-user-create-" + fixture.tenant.id))
	operationID := uuid.MustParse([]string{
		"59000000-0000-4000-8000-000000000001",
		"59000000-0000-4000-8000-000000000002",
	}[tenantIndex])
	for attempt := 0; attempt < 2; attempt++ {
		transaction := beginProfileRuntimeTransaction(
			t, ctx, database, fixture.tenant.schema,
		)
		var resultID uuid.UUID
		var revision, clock int64
		var replayed bool
		err := transaction.QueryRowContext(ctx, `SELECT * FROM
iam_identity_user_create_v2(
 pg_catalog.current_schema(),$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,
 $11,$12,$13,$14,$15,$16,$17,1,1)`,
			fixture.tenant.id, operationID, requestSHA[:],
			maintenanceLiveUser, maintenanceLiveIdentifier,
			maintenanceLiveAccessMode, maintenanceLiveAssignment,
			fixture.targetSchema.ID(), []byte(document), profileSHA[:],
			email, email, emailSHA[:], "directory-hmac-v1",
			identifierHMAC[:], maintenanceUserThree, maintenanceReadPermission,
		).Scan(&resultID, &revision, &clock, &replayed)
		if err == nil {
			err = transaction.Commit()
		} else {
			_ = transaction.Rollback()
		}
		if err != nil || resultID != uuid.MustParse(maintenanceLiveUser) ||
			revision != 1 || clock < 1 || replayed != (attempt == 1) {
			t.Fatalf("%s live user attempt %d=%s/%d/%d/%t: %v",
				fixture.tenant.schema, attempt, resultID,
				revision, clock, replayed, err)
		}
	}
	var schemaID uuid.UUID
	var revision int64
	var storedSHA []byte
	var receipts, assignments int
	err := database.QueryRowContext(ctx, `SELECT usr.profile_schema_version_id,
 usr.revision,usr.profile_sha256,
 (SELECT pg_catalog.count(*) FROM `+fixture.tenant.schema+
		`.iam_profile_backfill_preconversions AS receipt
   WHERE receipt.resource_id=usr.id),
 (SELECT pg_catalog.count(*) FROM `+fixture.tenant.schema+
		`.iam_user_role_assignments AS assignment
   WHERE assignment.user_id=usr.id AND assignment.role_id=$2)
FROM `+fixture.tenant.schema+`.iam_users AS usr WHERE usr.id=$1`,
		maintenanceLiveUser, maintenanceDefaultRole).Scan(
		&schemaID, &revision, &storedSHA, &receipts, &assignments,
	)
	if err != nil || schemaID != fixture.targetSchema.ID() || revision != 1 ||
		!bytes.Equal(storedSHA, profileSHA[:]) || receipts != 0 || assignments != 1 {
		t.Fatalf("%s live user state=%s/%d/%x/%d/%d: %v",
			fixture.tenant.schema, schemaID, revision, storedSHA,
			receipts, assignments, err)
	}
}

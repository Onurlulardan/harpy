//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestBillingInstallationCapabilityBackfillAgainstPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh", "v1_adopted"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()
			database := startBillingPG18(t, ctx, "capability_backfill_"+origin)
			prepareBillingPrerequisites(t, ctx, database, origin)
			documents := loadBillingDocuments(t, origin)
			hardening := billingMigrationIndex(
				t, documents, "20260729100007_",
			)
			capability := billingMigrationIndex(
				t, documents, "20260729100016_",
			)
			applyBillingDocuments(t, ctx, database, documents[:hardening+1])
			tenantID := billingV7(t)
			insertBillingTenantTopology(
				t, ctx, database, tenantID, 5, billingAcceptanceA,
			)
			operationID := billingV7(t)
			authorityID := billingInstallationIDFixture(
				operationID, "tenant-authority",
			)
			assignmentID := billingInstallationIDFixture(
				operationID, "base-offering-assignment",
			)
			offering := loadBillingControlOffering(t, ctx, database)
			assertBillingCapabilityBackfillPreservesLegacyABI(
				t, ctx, database, documents[capability],
				operationID, tenantID, offering,
			)
			insertLegacyBillingInstallation(t, ctx, database,
				operationID, tenantID, authorityID, assignmentID, offering)
			applyBillingDocuments(
				t, ctx, database, documents[hardening+1:capability+1],
			)
			var mode, kind, purpose string
			if err := database.QueryRowContext(ctx, `
SELECT deployment_mode,operation_kind,binding_purpose
FROM platform.billing_installation_assignment_contexts
WHERE operation_id=$1`, operationID).Scan(
				&mode, &kind, &purpose,
			); err != nil {
				t.Fatal(err)
			}
			if mode != "saas" || kind != "environment_initialize" ||
				purpose != "saas_control" {
				t.Fatalf("backfilled Billing context=%s/%s/%s",
					mode, kind, purpose)
			}
		})
	}
}

func billingMigrationIndex(
	t *testing.T,
	documents []billingDocument,
	prefix string,
) int {
	t.Helper()
	for index, document := range documents {
		if strings.HasPrefix(filepath.Base(document.path), prefix) {
			return index
		}
	}
	t.Fatalf("Billing migration prefix %s is unavailable", prefix)
	return -1
}

func loadBillingControlOffering(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) billingInstallationOffering {
	t.Helper()
	var offering billingInstallationOffering
	if err := database.QueryRowContext(ctx, `
SELECT offering_version_id,revision,offering_sha256,effective_at
FROM platform.billing_saas_control_offering_publications
WHERE offering_key='saas_platform_control_base_v1'`).Scan(
		&offering.versionID, &offering.revision,
		&offering.sha256, &offering.effective,
	); err != nil {
		t.Fatal(err)
	}
	return offering
}

func assertBillingCapabilityBackfillPreservesLegacyABI(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	document billingDocument,
	operationID, tenantID uuid.UUID,
	offering billingInstallationOffering,
) {
	t.Helper()
	transaction, err := database.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer transaction.Rollback()
	if _, err = transaction.ExecContext(ctx, `
SET SESSION AUTHORIZATION harpy_v2_runtime;
SET ROLE harpy_runtime`); err != nil {
		t.Fatal(err)
	}
	var replayed bool
	err = transaction.QueryRowContext(ctx, `
SELECT platform.billing_installation_assign_v2(
 $1,$2,$3,$4,$5,'saas',$6,$7,$8,$9,NULL,NULL,$10,'saas_control'
)`, operationID, uuid.MustParse(billingEnvironmentID), tenantID,
		billingV7(t), billingV7(t), offering.versionID, offering.revision,
		offering.sha256, offering.effective, billingAcceptanceA,
	).Scan(&replayed)
	if err != nil || replayed {
		t.Fatalf("seed Billing drift replay=%t err=%v", replayed, err)
	}
	if _, err = transaction.ExecContext(ctx,
		"RESET ROLE; RESET SESSION AUTHORIZATION"); err != nil {
		t.Fatal(err)
	}
	if _, err = transaction.ExecContext(ctx, string(document.content)); err != nil {
		t.Fatalf("Billing legacy-compatible expand: %v", err)
	}
	var contexts int
	if err = transaction.QueryRowContext(ctx, `
SELECT count(*) FROM platform.billing_installation_assignment_contexts
WHERE operation_id=$1`, operationID).Scan(&contexts); err != nil {
		t.Fatal(err)
	}
	if contexts != 0 {
		t.Fatalf("legacy Billing row received v3 context: %d", contexts)
	}
}

func insertLegacyBillingInstallation(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	operationID, tenantID, authorityID, assignmentID uuid.UUID,
	offering billingInstallationOffering,
) {
	t.Helper()
	connection, err := database.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()
	if _, err = connection.ExecContext(ctx, `
SET SESSION AUTHORIZATION harpy_v2_runtime;
SET ROLE harpy_runtime`); err != nil {
		t.Fatal(err)
	}
	var replayed bool
	err = connection.QueryRowContext(ctx, `
SELECT platform.billing_installation_assign_v2(
 $1,$2,$3,$4,$5,'saas',$6,$7,$8,$9,NULL,NULL,$10,'saas_control'
)`, operationID, uuid.MustParse(billingEnvironmentID), tenantID,
		authorityID, assignmentID, offering.versionID, offering.revision,
		offering.sha256, offering.effective, billingAcceptanceA,
	).Scan(&replayed)
	if err != nil || replayed {
		t.Fatalf("legacy Billing installation replay=%t err=%v", replayed, err)
	}
	if _, err = connection.ExecContext(ctx,
		"RESET ROLE; RESET SESSION AUTHORIZATION"); err != nil {
		t.Fatal(err)
	}
}

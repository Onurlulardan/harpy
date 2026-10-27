//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
)

type billingInstallationOffering struct {
	versionID uuid.UUID
	revision  int64
	sha256    string
	effective time.Time
}

func assertBillingInstallationUsesCurrentGeneration(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	tenantID := billingV7(t)
	insertBillingTenantTopology(
		t, ctx, database, tenantID, 5, billingAcceptanceA,
	)
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
	operationID := billingV7(t)
	authorityID := billingInstallationIDFixture(operationID, "tenant-authority")
	assignmentID := billingInstallationIDFixture(
		operationID, "base-offering-assignment",
	)
	if _, err := assignBillingInstallation(
		t, ctx, database, operationID, tenantID,
		authorityID, assignmentID, offering, false,
	); billingPGCode(err) != "42501" {
		t.Fatalf("installation spoof code=%q err=%v", billingPGCode(err), err)
	}
	replayed, err := assignBillingInstallation(
		t, ctx, database, operationID, tenantID,
		authorityID, assignmentID, offering, true,
	)
	if err != nil || replayed {
		t.Fatalf("current-generation installation = replay=%t err=%v", replayed, err)
	}
	var groupGeneration, billingGeneration, formsGeneration int64
	if err = database.QueryRowContext(ctx, `
SELECT authority_group_generation,billing_generation,forms_generation
FROM platform.billing_tenant_authorities
WHERE authority_id=$1`, authorityID).Scan(
		&groupGeneration, &billingGeneration, &formsGeneration,
	); err != nil {
		t.Fatal(err)
	}
	if groupGeneration != 5 || billingGeneration != 5 || formsGeneration != 5 {
		t.Fatalf("installation generations = %d/%d/%d, want 5/5/5",
			groupGeneration, billingGeneration, formsGeneration)
	}
	replayed, err = assignBillingInstallation(
		t, ctx, database, operationID, tenantID,
		authorityID, assignmentID, offering, true,
	)
	if err != nil || !replayed {
		t.Fatalf("installation replay = replay=%t err=%v", replayed, err)
	}
	setBillingAuthorityGeneration(
		t, ctx, database, tenantID.String(), "tenant", 6, billingAcceptanceB,
	)
	if _, err = assignBillingInstallation(
		t, ctx, database, operationID, tenantID,
		authorityID, assignmentID, offering, true,
	); billingPGCode(err) != "P5721" {
		t.Fatalf("stale installation replay code=%q err=%v", billingPGCode(err), err)
	}
	reconciled, remaining := reconcileBillingTenants(t, ctx, database, true)
	if reconciled != 1 || remaining != 0 {
		t.Fatalf("post-install successor reconcile=%d/%d", reconciled, remaining)
	}
}

func assignBillingInstallation(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	operationID, tenantID, authorityID, assignmentID uuid.UUID,
	offering billingInstallationOffering,
	mintCapability bool,
) (bool, error) {
	t.Helper()
	connection, err := database.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()
	transaction, err := connection.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer transaction.Rollback()
	if mintCapability {
		if err = mintBillingInstallationCommand(
			ctx, transaction, operationID, tenantID,
			offering, billingAcceptanceA,
		); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = transaction.ExecContext(ctx, `
SET SESSION AUTHORIZATION harpy_v2_runtime;
SET ROLE harpy_owner;
SELECT pg_catalog.set_config(
 'application_name','harpy-v2:core-operator-initialize',true
)`); err != nil {
		t.Fatal(err)
	}
	const query = `
SELECT platform.billing_installation_assign_v3(
 $1::uuid,$2::uuid,$3::uuid,$4::uuid,$5::uuid,$6::text,
 $7::uuid,$8::bigint,$9::text,$10::timestamptz,$11::bigint,
 $12::text,$13::text,$14::text
)`
	var replayed bool
	err = transaction.QueryRowContext(ctx, query,
		operationID, uuid.MustParse(billingEnvironmentID), tenantID,
		authorityID, assignmentID, "saas", offering.versionID,
		offering.revision, offering.sha256, offering.effective,
		nil, nil, billingAcceptanceA, "saas_control",
	).Scan(&replayed)
	if err != nil {
		return false, err
	}
	if _, err = transaction.ExecContext(ctx,
		"RESET ROLE; RESET SESSION AUTHORIZATION"); err != nil {
		return false, err
	}
	if err = transaction.Commit(); err != nil {
		return false, err
	}
	return replayed, err
}

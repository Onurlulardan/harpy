//go:build integration

package postgres

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
)

func assertInstallationAsyncEvidencePG18(
	t *testing.T, ctx context.Context, database iamPlatformGuardDatabase,
) {
	t.Helper()
	customerOperationID := uuid.MustParse("40000000-0000-4000-8000-000000000001")
	artifactID := uuid.MustParse("40000000-0000-4000-8000-000000000002")
	seedInstallationCapabilityPG18(t, ctx, database, customerOperationID,
		"activation.complete", "saas_customer_tenant_provision")
	if err := callInstallationEnqueuePG18(
		ctx, database, customerOperationID, artifactID,
	); err != nil {
		t.Fatalf("customer enqueue: %v", err)
	}
	if err := callInstallationEventPG18(
		ctx, database, customerOperationID,
	); err != nil {
		t.Fatalf("customer event: %v", err)
	}
	if !readInstallationEvidencePG18(
		t, ctx, database, customerOperationID, artifactID,
	) {
		t.Fatal("customer durable activation evidence is unavailable")
	}
	if readInstallationEvidencePG18(
		t, ctx, database, customerOperationID, uuid.New(),
	) {
		t.Fatal("customer evidence accepted the wrong artifact")
	}
	var receipts int
	if err := database.migrator.QueryRowContext(ctx, `
SELECT count(*)
FROM platform.platform_async_installation_activation_receipts
WHERE operation_id=$1`, customerOperationID).Scan(&receipts); err != nil {
		t.Fatal(err)
	}
	if receipts != 2 {
		t.Fatalf("customer activation receipts=%d", receipts)
	}

	environmentOperationID := uuid.MustParse("50000000-0000-4000-8000-000000000001")
	seedInstallationCapabilityPG18(t, ctx, database, environmentOperationID,
		"activation.complete", "environment_initialize")
	if err := callInstallationEventPG18(
		ctx, database, environmentOperationID,
	); err != nil {
		t.Fatalf("environment event: %v", err)
	}
	if !readInstallationEvidencePG18(
		t, ctx, database, environmentOperationID, uuid.Nil,
	) {
		t.Fatal("environment durable activation evidence is unavailable")
	}
	if readInstallationEvidencePG18(
		t, ctx, database, environmentOperationID, artifactID,
	) {
		t.Fatal("environment evidence accepted a nonexistent artifact")
	}
	if err := database.migrator.QueryRowContext(ctx, `
SELECT count(*)
FROM platform.platform_async_installation_activation_receipts
WHERE operation_id=$1`, environmentOperationID).Scan(&receipts); err != nil {
		t.Fatal(err)
	}
	if receipts != 1 {
		t.Fatalf("environment activation receipts=%d", receipts)
	}
	assertInstallationApplicationNamePG18(t, ctx, database)
}

func callInstallationEnqueuePG18(
	ctx context.Context, database iamPlatformGuardDatabase,
	operationID, artifactID uuid.UUID,
) error {
	artifactRef := fmt.Sprintf(
		"iam-lifecycle-delivery://tenant/%s/artifact/%s",
		installationPG18TenantID, artifactID,
	)
	var replay bool
	return database.runtime.QueryRowContext(ctx, `
SELECT platform.platform_async_initial_admin_setup_enqueue_v2(
    $1,$2,$3,$4,$5,$6,$7)`,
		operationID, installationPG18EnvironmentID, installationPG18TenantID,
		artifactID, artifactRef, installationPG18Schema,
		installationPG18Manifest).Scan(&replay)
}

func callInstallationEventPG18(
	ctx context.Context, database iamPlatformGuardDatabase, operationID uuid.UUID,
) error {
	var replay bool
	return database.runtime.QueryRowContext(ctx, `
SELECT platform.platform_async_installation_activation_event_v2(
    $1,$2,$3,$4,$5,$6)`,
		operationID, installationPG18EnvironmentID, installationPG18TenantID,
		installationPG18ActorID, installationPG18Schema,
		installationPG18Manifest).Scan(&replay)
}

func readInstallationEvidencePG18(
	t *testing.T, ctx context.Context, database iamPlatformGuardDatabase,
	operationID, artifactID uuid.UUID,
) bool {
	t.Helper()
	var expected any = artifactID
	if artifactID == uuid.Nil {
		expected = nil
	}
	var valid bool
	if err := database.runtime.QueryRowContext(ctx, `
SELECT platform.test_activation_evidence_proxy($1::uuid,$2::uuid)`,
		operationID, expected).Scan(&valid); err != nil {
		t.Fatal(err)
	}
	return valid
}

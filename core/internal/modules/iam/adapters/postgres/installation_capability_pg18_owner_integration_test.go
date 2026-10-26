//go:build integration

package postgres

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

var (
	installationPG18EnvironmentID = uuid.MustParse("10000000-0000-4000-8000-000000000001")
	installationPG18TenantID      = uuid.MustParse("10000000-0000-4000-8000-000000000002")
	installationPG18UserID        = uuid.MustParse("10000000-0000-4000-8000-000000000003")
	installationPG18ActorID       = uuid.MustParse("10000000-0000-4000-8000-000000000004")
)

const (
	installationPG18Schema   = "tenant_harpy"
	installationPG18Manifest = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" +
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
)

func assertInstallationNoCapabilityPG18(
	t *testing.T, ctx context.Context, database iamPlatformGuardDatabase,
) {
	t.Helper()
	operationID := uuid.MustParse("20000000-0000-4000-8000-000000000001")
	if err := callInstallationOwnerPG18(ctx, database, operationID); err == nil {
		t.Fatal("owner successor accepted a missing capability")
	}
	artifactID := uuid.MustParse("20000000-0000-4000-8000-000000000002")
	if err := callInstallationEnqueuePG18(
		ctx, database, operationID, artifactID,
	); err == nil {
		t.Fatal("enqueue successor accepted a missing capability")
	}
	if err := callInstallationEventPG18(ctx, database, operationID); err == nil {
		t.Fatal("event successor accepted a missing capability")
	}
	var writes int
	if err := database.migrator.QueryRowContext(ctx, `
SELECT (SELECT count(*) FROM platform.test_owner_writes)
     + (SELECT count(*) FROM platform.jobs)
     + (SELECT count(*) FROM platform.outbox)
     + (SELECT count(*)
        FROM platform.platform_async_installation_activation_receipts)`).
		Scan(&writes); err != nil {
		t.Fatal(err)
	}
	if writes != 0 {
		t.Fatalf("missing capability produced %d writes", writes)
	}
	_, err := database.runtime.ExecContext(ctx, `
INSERT INTO platform.platform_async_installation_activation_receipts(
    operation_id,receipt_kind,environment_id,tenant_id,event_id,manifest_sha256
) VALUES($1,'activation_event',$2,$3,$1,$4)`,
		operationID, installationPG18EnvironmentID,
		installationPG18TenantID, installationPG18Manifest)
	if err == nil {
		t.Fatal("runtime inserted an activation receipt directly")
	}
	var evidence bool
	err = database.runtime.QueryRowContext(ctx, `
SELECT platform.platform_async_installation_activation_evidence_assert_v1(
    $1::uuid,NULL::uuid)`, operationID).Scan(&evidence)
	if err == nil {
		t.Fatal("runtime executed the internal evidence assertion directly")
	}
	assertInstallationApplicationNamePG18(t, ctx, database)
}

func assertInstallationOwnerCapabilityPG18(
	t *testing.T, ctx context.Context, database iamPlatformGuardDatabase,
) {
	t.Helper()
	operationID := uuid.MustParse("30000000-0000-4000-8000-000000000001")
	seedInstallationCapabilityPG18(t, ctx, database, operationID,
		"platformauthority.bind_owner", "environment_initialize")
	if err := callInstallationOwnerPG18(ctx, database, operationID); err != nil {
		t.Fatalf("owner capability call: %v", err)
	}
	var writes int
	var consumed bool
	if err := database.migrator.QueryRowContext(ctx, `
SELECT (SELECT count(*) FROM platform.test_owner_writes WHERE operation_id=$1),
       (SELECT consumed FROM platform.test_activation_capabilities
        WHERE operation_id=$1 AND command_name='platformauthority.bind_owner')`,
		operationID).Scan(&writes, &consumed); err != nil {
		t.Fatal(err)
	}
	if writes != 1 || !consumed {
		t.Fatalf("owner writes=%d consumed=%t", writes, consumed)
	}
	if err := callInstallationOwnerPG18(ctx, database, operationID); err == nil {
		t.Fatal("consumed owner capability was reusable")
	}
	if err := database.migrator.QueryRowContext(ctx,
		"SELECT count(*) FROM platform.test_owner_writes WHERE operation_id=$1",
		operationID).Scan(&writes); err != nil {
		t.Fatal(err)
	}
	if writes != 1 {
		t.Fatalf("owner replay produced %d writes", writes)
	}
	assertInstallationApplicationNamePG18(t, ctx, database)
}

func callInstallationOwnerPG18(
	ctx context.Context, database iamPlatformGuardDatabase, operationID uuid.UUID,
) error {
	var replay bool
	return database.runtime.QueryRowContext(ctx, `
SELECT platform.iam_installation_instance_owner_bind_v2(
    $1,$2,$3,$4,$5,$6,$7,$8,$9,1,$10)`,
		operationID, installationPG18EnvironmentID, installationPG18TenantID,
		installationPG18Schema, installationPG18UserID, uuid.New(), uuid.New(),
		uuid.New(), uuid.New(), installationPG18Manifest).Scan(&replay)
}

func seedInstallationCapabilityPG18(
	t *testing.T, ctx context.Context, database iamPlatformGuardDatabase,
	operationID uuid.UUID, command, kind string,
) {
	t.Helper()
	_, err := database.migrator.ExecContext(ctx, `
INSERT INTO platform.test_activation_capabilities(
    operation_id,command_name,environment_id,tenant_id,schema_name,operation_kind
) VALUES($1,$2,$3,$4,$5,$6)`, operationID, command,
		installationPG18EnvironmentID, installationPG18TenantID,
		installationPG18Schema, kind)
	if err != nil {
		t.Fatal(err)
	}
}

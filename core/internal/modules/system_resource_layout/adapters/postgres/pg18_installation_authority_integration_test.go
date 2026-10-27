//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
	"time"

	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
)

const (
	layoutContractSchema               = "layout_contract"
	layoutContractOperationID          = "61000000-0000-4000-8000-000000000001"
	layoutContractTenantID             = "61000000-0000-4000-8000-000000000002"
	layoutContractEnvironmentID        = "61000000-0000-4000-8000-000000000003"
	layoutContractUserRevision         = "61000000-0000-7000-8000-000000000004"
	layoutContractOrganizationRevision = "61000000-0000-7000-8000-000000000005"
	layoutContractRoleRevision         = "61000000-0000-7000-8000-000000000006"
)

var layoutContractTime = time.Date(
	2026, time.July, 29, 12, 0, 0, 0, time.UTC,
)

func assertLayoutInstallationAuthorityContractPG18(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	prepareLayoutInstallationCapabilityPG18(t, ctx, database)
	const authorityContract = "20260729100047_contract_system_resource_layout_installation_authority_v1"
	gen70, gen71 := layoutMigrationRange(t, authorityContract, authorityContract)
	applyLayoutMigrationSetPG18(
		t, ctx, database, layoutContractSchema, gen70,
	)
	createAndPublishLayoutContractRevisionsPG18(t, ctx, database)
	if _, err := callLayoutInstallationRecordV1PG18(ctx, database); err != nil {
		t.Fatalf("Gen70 layout installation record: %v", err)
	}
	assertLayoutContractBaselineCountPG18(t, ctx, database, 1)
	applyLayoutMigrationSetPG18(t, ctx, database,
		layoutContractSchema, gen71)
	_, err := callLayoutInstallationRecordV1PG18(ctx, database)
	assertLayoutSQLState(t, err, "42501")
	assertLayoutContractBaselineCountPG18(t, ctx, database, 1)
	seedLayoutInstallationCapabilityPG18(t, ctx, database)
	if _, err := callLayoutInstallationRecordV2PG18(ctx, database); err != nil {
		t.Fatalf("Gen71 layout installation record: %v", err)
	}
	_, err = callLayoutInstallationRecordV2PG18(ctx, database)
	assertLayoutSQLState(t, err, "42501")
	assertLayoutContractBaselineCountPG18(t, ctx, database, 1)
	var consumed bool
	if err := database.QueryRowContext(ctx, `
SELECT consumed FROM platform.test_layout_capabilities
WHERE operation_id=$1`, layoutContractOperationID).Scan(&consumed); err != nil {
		t.Fatal(err)
	}
	if !consumed {
		t.Fatal("layout installation capability was not consumed")
	}
}

func createAndPublishLayoutContractRevisionsPG18(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	for _, revision := range []struct {
		id   string
		kind layoutapi.ResourceKind
	}{
		{layoutContractUserRevision, layoutapi.ResourceUser},
		{layoutContractOrganizationRevision, layoutapi.ResourceOrganization},
		{layoutContractRoleRevision, layoutapi.ResourceRole},
	} {
		createLayoutRevision(t, ctx, database, layoutContractSchema,
			revision.id, revision.kind, adapterDefinition)
		publishLayoutRevision(t, ctx, database, layoutContractSchema,
			revision.id, 1, adapterDefinition, layoutContractTime)
	}
}

func callLayoutInstallationRecordV1PG18(
	ctx context.Context, database *sql.DB,
) ([]byte, error) {
	return callLayoutInstallationRecordPG18(ctx, database, `
SELECT system_resource_layout_installation_record_v1(
 $1,$2,$3,$4,$5,$6)`)
}

func callLayoutInstallationRecordV2PG18(
	ctx context.Context, database *sql.DB,
) ([]byte, error) {
	return callLayoutInstallationRecordPG18(ctx, database, `
SELECT system_resource_layout_installation_record_v2(
 $1,$2,$3,$4,$5,$6,'layout_contract')`)
}

func callLayoutInstallationRecordPG18(
	ctx context.Context, database *sql.DB, statement string,
) ([]byte, error) {
	connection, transaction, err := beginLayoutRuntimeTx(
		ctx, database, layoutContractSchema,
	)
	if err != nil {
		return nil, err
	}
	defer connection.Close()
	var digest []byte
	err = transaction.QueryRowContext(ctx, statement,
		layoutContractOperationID, layoutContractTenantID,
		layoutContractUserRevision, layoutContractOrganizationRevision,
		layoutContractRoleRevision, layoutContractTime).Scan(&digest)
	if err != nil {
		_ = transaction.Rollback()
		return nil, err
	}
	if err := transaction.Commit(); err != nil {
		return nil, err
	}
	return digest, nil
}

func assertLayoutContractBaselineCountPG18(
	t *testing.T, ctx context.Context, database *sql.DB, expected int,
) {
	t.Helper()
	var count int
	query := `SELECT count(*) FROM layout_contract.` +
		`system_resource_layout_installation_baselines WHERE operation_id=$1`
	if err := database.QueryRowContext(ctx, query,
		layoutContractOperationID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != expected {
		t.Fatalf("layout installation baseline count = %d, want %d",
			count, expected)
	}
}

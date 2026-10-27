//go:build integration

package maintenance

import (
	"bytes"
	"context"
	"database/sql"
	"testing"
	"time"
)

const (
	maintenanceDirectIdentifier = "019c6a7e-7c00-7000-8000-000000000011"
	maintenanceDirectUser       = "019c6a7e-7c00-7000-8000-000000000012"
)

func assertMaintenancePG18DirectRuntime(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	assertMaintenancePG18EvidenceConstraints(t, ctx, database)
	connection, err := database.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()
	if _, err = connection.ExecContext(
		ctx, "SET SESSION AUTHORIZATION harpy_v2_runtime",
	); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_, _ = connection.ExecContext(context.Background(), "RESET SESSION AUTHORIZATION")
	}()
	if _, err = connection.ExecContext(
		ctx, "SET search_path=t_iam_a,platform,pg_catalog",
	); err != nil {
		t.Fatal(err)
	}
	assertMaintenancePG18DirectBootstrap(t, ctx, connection)
	assertMaintenancePG18DirectProjection(t, ctx, connection)
	if _, err = connection.ExecContext(ctx, "RESET SESSION AUTHORIZATION"); err != nil {
		t.Fatal(err)
	}
	assertMaintenancePG18Isolation(t, ctx, connection)
}

func assertMaintenancePG18DirectBootstrap(
	t *testing.T, ctx context.Context, connection *sql.Conn,
) {
	t.Helper()
	var reference string
	var digestLength int
	var created bool
	err := connection.QueryRowContext(ctx, `SELECT payload_reference,
 pg_catalog.octet_length(payload_sha256),created
FROM iam_maintenance_bootstrap_v1(
 pg_catalog.current_schema(),$1::uuid,'iam.directory.repair',$2::uuid,$3)`,
		maintenancePG18Tenants[0].id, iamMaintenanceJobID,
		time.Date(2026, 7, 16, 12, 0, 0, 0, time.UTC),
	).Scan(&reference, &digestLength, &created)
	if err != nil || reference == "" || digestLength != 32 || !created {
		t.Fatalf("direct bootstrap=%q/%d/%t: %v",
			reference, digestLength, created, err)
	}
}

func assertMaintenancePG18DirectProjection(
	t *testing.T, ctx context.Context, connection *sql.Conn,
) {
	t.Helper()
	sha := bytes.Repeat([]byte{0x31}, 32)
	evidence := `[{"key_id":"AAAAAAAAAAAAAAAA","hmac":"` +
		`0000000000000000000000000000000000000000000000000000000000000000"}]`
	const apply = `SELECT platform.iam_maintenance_directory_apply_v1(
 $1::uuid,$2::uuid,$3::uuid,'email_v1',$4,$5,$6::bytea,$7::jsonb)`
	var affected int64
	err := connection.QueryRowContext(ctx, apply,
		maintenancePG18Tenants[0].id, maintenanceDirectIdentifier,
		maintenanceDirectUser, "active", 1, sha, evidence).Scan(&affected)
	if err != nil || affected != 1 {
		t.Fatalf("direct projection=%d/%v", affected, err)
	}
	assertMaintenancePG18SnapshotCount(
		t, ctx, connection, maintenancePG18Tenants[0].id, 1,
	)
	assertMaintenancePG18SnapshotCount(
		t, ctx, connection, maintenancePG18Tenants[1].id, 0,
	)
	err = connection.QueryRowContext(ctx, apply,
		maintenancePG18Tenants[0].id, maintenanceDirectIdentifier,
		maintenanceDirectUser, "retired", 2, sha, `[]`).Scan(&affected)
	if err != nil || affected != 1 {
		t.Fatalf("direct tombstone=%d/%v", affected, err)
	}
	if _, err = connection.ExecContext(ctx, "RESET SESSION AUTHORIZATION"); err != nil {
		t.Fatal(err)
	}
	_, err = connection.ExecContext(ctx, `UPDATE platform.iam_login_directory
SET projected_at=pg_catalog.transaction_timestamp()-INTERVAL '31 days',
 tombstoned_at=pg_catalog.transaction_timestamp()-INTERVAL '31 days'
WHERE tenant_id=$1 AND source_identifier_id=$2`,
		maintenancePG18Tenants[0].id, maintenanceDirectIdentifier)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = connection.ExecContext(
		ctx, "SET SESSION AUTHORIZATION harpy_v2_runtime",
	); err != nil {
		t.Fatal(err)
	}
	err = connection.QueryRowContext(ctx,
		"SELECT platform.iam_maintenance_directory_purge_v1($1,$2,$3)",
		maintenancePG18Tenants[0].id, maintenanceDirectIdentifier, 2,
	).Scan(&affected)
	if err != nil || affected != 1 {
		t.Fatalf("direct purge=%d/%v", affected, err)
	}
	assertMaintenancePG18SnapshotCount(
		t, ctx, connection, maintenancePG18Tenants[0].id, 0,
	)
}

func assertMaintenancePG18SnapshotCount(
	t *testing.T, ctx context.Context, connection *sql.Conn,
	tenantID string, want int,
) {
	t.Helper()
	var count int
	err := connection.QueryRowContext(ctx, `SELECT pg_catalog.count(*)
FROM platform.iam_maintenance_directory_snapshot_v1($1::uuid)`, tenantID).Scan(&count)
	if err != nil || count != want {
		t.Fatalf("snapshot %s=%d/%v, want %d", tenantID, count, err, want)
	}
}

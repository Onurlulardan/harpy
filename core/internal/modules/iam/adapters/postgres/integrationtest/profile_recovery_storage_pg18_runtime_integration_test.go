//go:build integration

package integrationtest

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func assertRecoveryRuntimeRoleBoundary(
	t *testing.T, ctx context.Context, db publicationDatabase,
) {
	t.Helper()
	expectRecoveryRuntimeError(t, ctx, db, false,
		`SELECT t_profile_publication.iam_profile_read_contract_v2(
't_profile_publication','user')`, "42501")
	expectRecoveryRuntimeError(t, ctx, db, true,
		`SELECT iam_profile_read_contract_v1('t_profile_publication','user')`, "42501")
	expectRecoveryRuntimeError(t, ctx, db, true,
		`SELECT count(*) FROM iam_profile_recovery_resources_v1`, "42501")
	conn := openRecoveryRuntimeConnection(t, ctx, db)
	tx := beginRecoveryRuntimeRole(t, ctx, conn, db)
	var contract []byte
	if err := tx.QueryRow(ctx, `SELECT iam_profile_read_contract_v2($1,'user')`,
		db.tenant.schema).Scan(&contract); err != nil {
		t.Fatalf("call recovery read contract with runtime role: %v", err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
}

func assertRecoveryMixedInventory(
	t *testing.T, ctx context.Context, db publicationDatabase, parentID uuid.UUID,
) {
	t.Helper()
	conn := openRecoveryRuntimeConnection(t, ctx, db)
	tx := beginRecoveryRuntimeRole(t, ctx, conn, db)
	var cursor string
	err := tx.QueryRow(ctx, `SELECT iam_profile_recovery_inventory_open_v1(
$1,$2,$3,$4,1,1)`,
		db.tenant.schema, parentID, publicationActorID, publicationManagePerm).
		Scan(&cursor)
	if err != nil {
		t.Fatalf("open mixed recovery inventory: %v", err)
	}
	rows, err := tx.Query(ctx,
		"FETCH FORWARD 64 FROM "+pgx.Identifier{cursor}.Sanitize())
	if err != nil {
		t.Fatalf("read mixed recovery inventory: %v", err)
	}
	counts := map[string]int{}
	for rows.Next() {
		var resourceID, schemaID uuid.UUID
		var version, revision int64
		var state, lineage, document string
		var digest []byte
		if err = rows.Scan(&resourceID, &schemaID, &version, &state, &lineage,
			&revision, &document, &digest); err != nil {
			t.Fatal(err)
		}
		counts[lineage]++
	}
	rows.Close()
	if err = rows.Err(); err != nil || counts["source"] != 1 || counts["target"] != 1 {
		t.Fatalf("mixed recovery inventory = %#v/%v", counts, err)
	}
	if _, err = tx.Exec(ctx, "CLOSE "+pgx.Identifier{cursor}.Sanitize()); err != nil {
		t.Fatalf("close mixed recovery inventory: %v", err)
	}
	var contract []byte
	if err = tx.QueryRow(ctx,
		`SELECT iam_profile_read_contract_v2($1,'user')`, db.tenant.schema).
		Scan(&contract); err != nil {
		t.Fatalf("read fenced mixed contract: %v", err)
	}
	var decoded struct {
		PlanID   string            `json:"plan_id"`
		Target   json.RawMessage   `json:"target"`
		Mappings []json.RawMessage `json:"mappings"`
	}
	if err = json.Unmarshal(contract, &decoded); err != nil ||
		decoded.PlanID != parentID.String() || string(decoded.Target) == "null" ||
		len(decoded.Mappings) != 2 {
		t.Fatalf("fenced mixed contract = %s/%v", contract, err)
	}
	if err = tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
}

func assertRecoveryFencedWriteContracts(
	t *testing.T, ctx context.Context, db publicationDatabase,
) {
	t.Helper()
	expectRecoveryRuntimeError(t, ctx, db, true,
		`SELECT iam_profile_write_contract_read_v2(
't_profile_publication','user')`, "P5741")
	installRecoveryFenceProbe(t, ctx, db)
	expectRecoveryRuntimeError(t, ctx, db, true,
		`SELECT iam_profile_recovery_fence_probe_v1(
'51000000-0000-4000-8000-000000000001')`, "P5741")
}

func installRecoveryFenceProbe(
	t *testing.T, ctx context.Context, db publicationDatabase,
) {
	t.Helper()
	const statement = `CREATE FUNCTION iam_profile_recovery_fence_probe_v1(p_id uuid)
RETURNS void LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$ BEGIN
UPDATE iam_users SET profile_sha256=pg_catalog.sha256(
pg_catalog.convert_to('fenced-probe','UTF8')),revision=revision+1,
updated_at=pg_catalog.now() WHERE id=p_id;
END; $function$;
REVOKE ALL ON FUNCTION iam_profile_recovery_fence_probe_v1(uuid) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION iam_profile_recovery_fence_probe_v1(uuid)
TO harpy_runtime`
	execPublicationTenant(t, ctx, db, statement)
}

func openRecoveryRuntimeConnection(
	t *testing.T, ctx context.Context, db publicationDatabase,
) *pgx.Conn {
	t.Helper()
	conn, err := pgx.Connect(ctx, db.runtimeDSN)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close(context.Background()) })
	return conn
}

func beginRecoveryRuntimeRole(
	t *testing.T, ctx context.Context, conn *pgx.Conn, db publicationDatabase,
) pgx.Tx {
	t.Helper()
	tx, err := conn.Begin(ctx)
	if err == nil {
		_, err = tx.Exec(ctx, "SET LOCAL ROLE harpy_runtime")
	}
	if err == nil {
		_, err = tx.Exec(ctx, "SELECT pg_catalog.set_config('search_path',$1,true)",
			db.tenant.schema+",platform,pg_catalog")
	}
	if err != nil {
		_ = tx.Rollback(ctx)
		t.Fatal(err)
	}
	return tx
}

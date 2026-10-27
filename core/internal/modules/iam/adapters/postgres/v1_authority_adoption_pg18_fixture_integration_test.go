//go:build integration

package postgres

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"testing"
)

const (
	v1AdoptionEnvironment = "71000000-0000-4000-8000-000000000001"
	v1AdoptionTenant      = "71000000-0000-4000-8000-000000000002"
	v1AdoptionUser        = "71000000-0000-4000-8000-000000000003"
	v1AdoptionRole        = "71000000-0000-4000-8000-000000000004"
	v1AdoptionRun         = "71000000-0000-4000-8000-000000000010"
)

const v1AdoptionPlan = `{"environment_id":"71000000-0000-4000-8000-000000000001","mode":"onpremise","primary_tenant_id":"71000000-0000-4000-8000-000000000002","owner_user_id":"71000000-0000-4000-8000-000000000003","owner_authorization_version":7,"owner_mapping_sha256":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","tenant_bindings":[],"base_assignments":[],"inventory_sha256":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","plan_sha256":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"}`

var v1AuthorityPlatformMigrations = []string{
	"20260716010000_create_iam_login_directory",
	"20260722020000_create_iam_platform_catalog",
	"20260722020001_create_iam_platform_authority",
	"20260722020002_create_iam_platform_policy",
	"20260722020003_guard_iam_platform_authority",
	"20260722020004_guard_iam_platform_policy",
	"20260722020005_guard_iam_platform_catalog",
	"20260722020006_seed_iam_platform_catalog",
	"20260722020007_seed_iam_platform_permissions",
	"20260723111000_create_iam_v1_protected_authority_receipts_v1",
	"20260723111001_create_iam_v1_adopt_protected_authority_v1",
	"20260723111003_revoke_direct_iam_v1_protected_authority_v1",
	"20260723111004_create_iam_v1_protected_authority_apply_v1",
	"20260723111005_create_iam_v1_protected_authority_composite_v1",
}

func prepareV1AuthorityAdoption(
	t *testing.T, ctx context.Context, database iamPlatformGuardDatabase, origin string,
) {
	t.Helper()
	if _, err := database.admin.ExecContext(ctx, `GRANT harpy_owner
TO harpy_adopter WITH INHERIT FALSE, SET TRUE`); err != nil {
		t.Fatalf("grant adoption owner role: %v", err)
	}
	connection, err := database.migrator.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()
	if _, err = connection.ExecContext(ctx, v1AuthorityAdoptionSchema); err != nil {
		t.Fatalf("prepare V1 authority adoption schema: %v", err)
	}
	for _, name := range v1AuthorityPlatformMigrations {
		applyIAMPlatformGuardDocument(t, ctx, connection, name,
			readIAMPlatformAuthMigration(t, origin, name))
	}
	if _, err = connection.ExecContext(ctx, v1AuthorityDirectoryEvidence); err != nil {
		t.Fatalf("seed V1 authority directory evidence: %v", err)
	}
	if _, err = connection.ExecContext(ctx, `INSERT INTO
platform.platform_cutover_v1_adoption_plans VALUES (
$1,$2,'ready',$3,$4,repeat('b',64),$5::jsonb)`,
		v1AdoptionOperation, v1AdoptionEnvironment, v1AdoptionTenant,
		v1AdoptionUser, v1AdoptionPlan); err != nil {
		t.Fatalf("seed V1 authority adoption plan: %v", err)
	}
	if _, err = connection.ExecContext(ctx,
		"GRANT USAGE ON SCHEMA platform TO harpy_adopter"); err != nil {
		t.Fatalf("grant adoption schema use: %v", err)
	}
}

func callV1AuthorityAdoption(
	ctx context.Context, adopter *sql.DB, batch []byte, digest string,
) (int64, error) {
	var count int64
	err := adopter.QueryRowContext(ctx, `SELECT applied_count FROM
platform.iam_v1_protected_authority_apply_and_record_v1(
$1::uuid,$2::uuid,'global'::text,$3::uuid,1::smallint,
'v1-authority'::text,'snapshot'::text,'0/10'::pg_lsn,
'0/10'::pg_lsn,'0/10'::pg_lsn,'0/10'::pg_lsn,'10'::text::xid8,
1::integer,'complete'::text,$4::jsonb,$5::bytea,$6::text,$7::bigint,
repeat('c',64)::text)`,
		v1AdoptionOperation, v1AdoptionRun, v1AdoptionEnvironment,
		`{"operation_id":"`+v1AdoptionOperation+`"}`,
		batch, digest, len(batch)).Scan(&count)
	return count, err
}

func v1AuthorityAdoptionBatch(t *testing.T) ([]byte, string) {
	t.Helper()
	after := map[string]any{
		"contract_version": "platform-v1-environment-adoption.v1",
		"operation_id":     v1AdoptionOperation,
		"inventory":        map[string]any{},
		"plan":             json.RawMessage(v1AdoptionPlan),
		"authority_ids": map[string]string{
			"subject_id":         "71000000-0000-4000-8000-000000000005",
			"owner_role_id":      "71000000-0000-4000-8000-000000000006",
			"role_permission_id": "71000000-0000-4000-8000-000000000007",
			"owner_binding_id":   "71000000-0000-4000-8000-000000000008",
		},
	}
	batch := map[string]any{
		"operation_id": v1AdoptionOperation, "run_id": v1AdoptionRun,
		"scope_kind": "global", "scope_id": v1AdoptionEnvironment,
		"relation_ordinal": 1, "checkpoint_key": "v1-authority",
		"checkpoint_state": "complete",
		"last_key":         map[string]string{"operation_id": v1AdoptionOperation},
		"source_position": map[string]any{
			"snapshot_name": "snapshot", "snapshot_lsn": "0/10",
			"received_lsn": "0/10", "commit_lsn": "0/10",
			"final_lsn": "0/10", "source_xid": "10", "batch_ordinal": 1,
		},
		"changes": []any{map[string]any{
			"ordinal": 1, "kind": "insert",
			"key":    map[string]string{"operation_id": v1AdoptionOperation},
			"before": nil, "after": after,
		}},
	}
	encoded, err := json.Marshal(batch)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(encoded)
	return encoded, hex.EncodeToString(digest[:])
}

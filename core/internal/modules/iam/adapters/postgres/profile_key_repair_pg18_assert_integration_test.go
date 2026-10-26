//go:build integration

package postgres

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"testing"
)

func applyIAMProfileRepairTargetsPG18(
	t *testing.T, ctx context.Context, database *sql.DB, origin string,
	schemas ...string,
) {
	t.Helper()
	names := []string{
		"20260730120000_repair_iam_profile_keys_and_initialize_v3",
		"20260730120001_create_iam_installation_initialize_v3",
		"20260730120002_normalize_iam_legacy_profile_writes_v1",
	}
	documents := make([]iamFoundationDocument, 0, len(names))
	for _, name := range names {
		documents = append(documents, iamFoundationDocument{
			path: name, content: readIAMIdentityMigration(t, origin, name),
		})
	}
	for _, schema := range schemas {
		applyIAMTenantDocuments(t, ctx, database, schema, documents)
	}
}

func assertIAMProfileRepairRejectsCorruptionPG18(
	t *testing.T, ctx context.Context, database iamPlatformGuardDatabase,
	origin string, fixtures iamProfileRepairFixtures,
) {
	t.Helper()
	corruptIAMProfileRepairPG18(
		t, ctx, database.migrator, fixtures.badShape.schema,
		`UPDATE iam_users SET
 profile_values='{"display_name":"Invalid","extra":"value"}'::jsonb,
 profile_sha256=pg_catalog.sha256(pg_catalog.convert_to(
  '{"display_name":"Invalid","extra":"value"}'::jsonb::text,'UTF8')),
 revision=revision+1,updated_at=pg_catalog.statement_timestamp()
WHERE id=$1`, fixtures.badShape.user,
	)
	corruptIAMProfileRepairPG18(
		t, ctx, database.migrator, fixtures.badDigest.schema,
		`UPDATE iam_users SET
 profile_sha256=pg_catalog.decode(pg_catalog.repeat('ff',32),'hex'),
 revision=revision+1,updated_at=pg_catalog.statement_timestamp()
WHERE id=$1`, fixtures.badDigest.user,
	)
	for _, fixture := range []iamProfileRepairInstallation{
		fixtures.badShape, fixtures.badDigest,
	} {
		before := snapshotIAMProfileRepairPG18(
			t, ctx, database.migrator, fixture.schema,
		)
		document := readIAMIdentityMigration(t, origin,
			"20260730120000_repair_iam_profile_keys_and_initialize_v3")
		connection, transaction := beginIAMTenantTx(
			t, ctx, database.migrator, fixture.schema,
		)
		_, err := transaction.ExecContext(ctx, string(document))
		_ = transaction.Rollback()
		connection.Close()
		assertIAMSQLState(t, err, "23514")
		after := snapshotIAMProfileRepairPG18(
			t, ctx, database.migrator, fixture.schema,
		)
		if after != before {
			t.Fatalf("%s profile repair rejection was not atomic", fixture.schema)
		}
	}
}

func corruptIAMProfileRepairPG18(
	t *testing.T, ctx context.Context, database *sql.DB,
	schema, statement, userID string,
) {
	t.Helper()
	connection, transaction := beginIAMTenantTx(t, ctx, database, schema)
	defer connection.Close()
	if _, err := transaction.ExecContext(ctx, statement, userID); err != nil {
		_ = transaction.Rollback()
		t.Fatalf("corrupt %s profile fixture: %v", schema, err)
	}
	if err := transaction.Commit(); err != nil {
		t.Fatal(err)
	}
}

func snapshotIAMProfileRepairPG18(
	t *testing.T, ctx context.Context, database *sql.DB, schema string,
) string {
	t.Helper()
	var snapshot string
	mustQueryIAMTenantValue(t, ctx, database, schema, `
SELECT pg_catalog.string_agg(
 row.kind||':'||row.id::text||':'||row.profile_values::text||':'||
 pg_catalog.encode(row.profile_sha256,'hex')||':'||row.revision::text,
 '|' ORDER BY row.kind)
FROM (
 SELECT 'organization' AS kind,id,profile_values,profile_sha256,revision
 FROM iam_organizations
 UNION ALL
 SELECT 'role',id,profile_values,profile_sha256,revision FROM iam_roles
 UNION ALL
 SELECT 'user',id,profile_values,profile_sha256,revision FROM iam_users
) AS row`, &snapshot)
	return snapshot
}

func assertIAMProfileRepairRowsPG18(
	t *testing.T, ctx context.Context, database *sql.DB,
	fixture iamProfileRepairInstallation, revision int64,
) {
	t.Helper()
	rows := []struct{ table, id, field, value string }{
		{"iam_users", fixture.user, fixture.userField, fixture.displayName},
		{"iam_organizations", fixture.organization,
			fixture.organizationField, fixture.tenantName},
		{"iam_roles", fixture.role, fixture.roleField, "Protected Tenant Admin"},
	}
	for _, expected := range rows {
		connection, transaction := beginIAMTenantTx(
			t, ctx, database, fixture.schema,
		)
		var profileJSON string
		var digest []byte
		var gotRevision int64
		err := transaction.QueryRowContext(ctx,
			`SELECT profile_values::text,profile_sha256,revision FROM `+
				expected.table+` WHERE id=$1`, expected.id).
			Scan(&profileJSON, &digest, &gotRevision)
		_ = transaction.Rollback()
		connection.Close()
		if err != nil {
			t.Fatal(err)
		}
		var values map[string]string
		if err := json.Unmarshal([]byte(profileJSON), &values); err != nil {
			t.Fatal(err)
		}
		if len(values) != 1 || values[expected.field] != expected.value {
			t.Fatalf("%s profile = %s", expected.table, profileJSON)
		}
		canonical, err := json.Marshal(map[string]string{
			expected.field: expected.value,
		})
		if err != nil {
			t.Fatal(err)
		}
		wantDigest := sha256.Sum256(canonical)
		if !bytes.Equal(digest, wantDigest[:]) || gotRevision != revision {
			t.Fatalf("%s digest/revision mismatch: %x/%d want %x/%d",
				expected.table, digest, gotRevision, wantDigest, revision)
		}
	}
}

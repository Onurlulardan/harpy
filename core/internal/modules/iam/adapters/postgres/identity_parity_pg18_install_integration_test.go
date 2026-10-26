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

func applyIAMIdentityParityInstallTargetsPG18(
	t *testing.T, ctx context.Context, database *sql.DB, origin string,
	schemas ...string,
) {
	t.Helper()
	documents := iamProfileRepairDocuments(t, origin, []string{
		"20260801015100_create_iam_installation_derived_id_v1",
		"20260801015101_upgrade_iam_installation_user_profile_v1",
		"20260801015102_create_iam_installation_initialize_v4",
	})
	for _, schema := range schemas {
		applyIAMTenantDocuments(t, ctx, database, schema, documents)
	}
}

func assertIAMInstallationUserProfileV2PG18(
	t *testing.T, ctx context.Context, database *sql.DB,
	fixture iamProfileRepairInstallation, revision int64,
) {
	t.Helper()
	targetID := iamInstallationDerivedID(
		fixture.operation, "user-profile-schema-v2",
	)
	displayID := iamInstallationDerivedID(
		fixture.operation, "user-display-name-field",
	)
	firstID := iamInstallationDerivedID(
		fixture.operation, "user-first-name-field",
	)
	lastID := iamInstallationDerivedID(
		fixture.operation, "user-last-name-field",
	)
	phoneID := iamInstallationDerivedID(
		fixture.operation, "user-phone-field",
	)
	connection, transaction := beginIAMTenantTx(
		t, ctx, database, fixture.schema,
	)
	defer connection.Close()
	var fields int
	err := transaction.QueryRowContext(ctx, `
SELECT count(*) FROM iam_profile_schema_versions AS schema
JOIN iam_profile_field_revisions AS field
  ON field.schema_version_id=schema.id
WHERE schema.id=$1 AND schema.state='active'
  AND field.field_state='active' AND field.classification='public'
  AND field.field_id IN ($2,$3,$4,$5)`,
		targetID, displayID, firstID, lastID, phoneID).Scan(&fields)
	if err != nil || fields != 4 {
		_ = transaction.Rollback()
		t.Fatalf("%s successor public fields=%d", fixture.schema, fields)
	}
	var raw string
	var digest []byte
	var actualRevision int64
	err = transaction.QueryRowContext(ctx, `
SELECT profile_values::text,profile_sha256,revision
FROM iam_users WHERE id=$1 AND profile_schema_version_id=$2`,
		fixture.user, targetID).Scan(&raw, &digest, &actualRevision)
	_ = transaction.Rollback()
	if err != nil {
		t.Fatal(err)
	}
	values := map[string]string{}
	if err := json.Unmarshal([]byte(raw), &values); err != nil {
		t.Fatal(err)
	}
	if len(values) != 2 || values[displayID] != fixture.displayName ||
		values[firstID] != fixture.displayName || values[lastID] != "" ||
		values[phoneID] != "" {
		t.Fatalf("%s successor profile=%s", fixture.schema, raw)
	}
	canonical, _ := json.Marshal(values)
	wantDigest := sha256.Sum256(canonical)
	if !bytes.Equal(digest, wantDigest[:]) || actualRevision != revision {
		t.Fatalf("%s digest/revision=%x/%d want=%x/%d",
			fixture.schema, digest, actualRevision, wantDigest, revision)
	}
}

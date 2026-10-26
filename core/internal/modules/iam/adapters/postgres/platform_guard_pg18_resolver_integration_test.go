//go:build integration

package postgres

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

func assertIAMTargetResolvers(
	t *testing.T, ctx context.Context, database iamPlatformGuardDatabase,
	refreshSelector []byte,
) {
	t.Helper()
	hmac := iamBytes(41, 32)
	seedIAMLoginTarget(t, ctx, database, iamGuardTenantA, hmac, 4)
	assertIAMTenantTarget(t, ctx, database, 1)
	assertIAMLoginTarget(t, ctx, database, iamGuardTenantA, hmac, 4, 1)
	assertIAMLoginCrossTenantIsolation(t, ctx, database, hmac)
	assertIAMRefreshTarget(t, ctx, database, refreshSelector, 2, 1)
	assertIAMExpiredRefreshTarget(t, ctx, database)
}

func seedIAMLoginTarget(
	t *testing.T, ctx context.Context, database iamPlatformGuardDatabase,
	tenantID string, hmac []byte, revision int64,
) {
	t.Helper()
	_, err := database.migrator.ExecContext(ctx, `
INSERT INTO platform.iam_login_directory(
 tenant_id,identifier_kind,lookup_key_id,identifier_hmac,user_id,
 source_identifier_id,source_revision,state,source_sha256,projected_at,tombstoned_at)
VALUES($1,'email_v1','guard-key',$2,$3,$4,$5,'active',$6,$7,NULL)`,
		tenantID, hmac, uuid.New(), uuid.New(), revision,
		iamBytes(byte(revision), 32), time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
}

func assertIAMTenantTarget(
	t *testing.T, ctx context.Context, database iamPlatformGuardDatabase,
	generation int64,
) {
	t.Helper()
	var tenant string
	var group, iam, layout int64
	err := database.runtime.QueryRowContext(ctx, `SELECT * FROM
platform.iam_resolve_tenant_target_v1($1)`, iamGuardTenantA).
		Scan(&tenant, &group, &iam, &layout)
	if err != nil || tenant != iamGuardTenantA || group != generation ||
		iam != generation || layout != generation {
		t.Fatalf("tenant target=%s/%d/%d/%d err=%v",
			tenant, group, iam, layout, err)
	}
}

func assertIAMLoginTarget(
	t *testing.T, ctx context.Context, database iamPlatformGuardDatabase,
	tenantID string, hmac []byte, wantRevision, generation int64,
) {
	t.Helper()
	var tenant string
	var revision, group, iam, layout int64
	var sourceHash []byte
	var userID, identifierID string
	var identifierKind string
	err := database.runtime.QueryRowContext(ctx, `SELECT * FROM
platform.iam_resolve_login_target_v2($1,'email_v1','guard-key',$2)`, tenantID, hmac).
		Scan(&tenant, &userID, &identifierID, &identifierKind,
			&revision, &sourceHash, &group, &iam, &layout)
	if err != nil || tenant != tenantID || revision != wantRevision ||
		uuid.Validate(userID) != nil || uuid.Validate(identifierID) != nil ||
		identifierKind != "email_v1" ||
		!bytes.Equal(sourceHash, iamBytes(byte(wantRevision), 32)) ||
		group != generation || iam != generation || layout != generation {
		t.Fatalf("unexpected login target tenant=%s revision=%d gens=%d/%d/%d err=%v",
			tenant, revision, group, iam, layout, err)
	}
}

func assertIAMLoginCrossTenantIsolation(
	t *testing.T, ctx context.Context, database iamPlatformGuardDatabase, hmac []byte,
) {
	t.Helper()
	seedIAMLoginTarget(t, ctx, database, iamGuardTenantB, hmac, 5)
	assertIAMLoginTarget(t, ctx, database, iamGuardTenantA, hmac, 4, 1)
	assertIAMLoginTarget(t, ctx, database, iamGuardTenantB, hmac, 5, 1)
	if _, err := database.migrator.ExecContext(ctx, `DELETE FROM
platform.iam_login_directory WHERE tenant_id=$1 AND lookup_key_id='guard-key'`,
		iamGuardTenantB); err != nil {
		t.Fatal(err)
	}
	err := database.runtime.QueryRowContext(ctx, `SELECT * FROM
platform.iam_resolve_login_target_v2($1,'email_v1','guard-key',$2)`,
		iamGuardTenantB, hmac).
		Scan(new(string), new(string), new(string), new(string), new(int64),
			new([]byte), new(int64), new(int64), new(int64))
	assertIAMSQLState(t, err, "P5731")
}

func assertIAMRefreshTarget(
	t *testing.T, ctx context.Context, database iamPlatformGuardDatabase,
	selector []byte, wantRevision, generation int64,
) {
	t.Helper()
	var tenant, tokenID, familyID string
	var revision, group, iam, layout int64
	var sourceHash, storedSelector []byte
	err := database.runtime.QueryRowContext(ctx, `SELECT * FROM
platform.iam_resolve_refresh_target_v2($1)`, selector).
		Scan(&tenant, &tokenID, &familyID, &storedSelector, &revision,
			&sourceHash, &group, &iam, &layout)
	if err != nil || tenant != iamGuardTenantA || revision != wantRevision ||
		uuid.Validate(tokenID) != nil || uuid.Validate(familyID) != nil ||
		!bytes.Equal(storedSelector, selector) ||
		!bytes.Equal(sourceHash, iamBytes(byte(wantRevision), 32)) ||
		group != generation || iam != generation || layout != generation {
		t.Fatalf("unexpected refresh target tenant=%s revision=%d err=%v",
			tenant, revision, err)
	}
}

func assertIAMExpiredRefreshTarget(
	t *testing.T, ctx context.Context, database iamPlatformGuardDatabase,
) {
	t.Helper()
	now := time.Now().UTC()
	selector := iamBytes(42, 16)
	err := createIAMGuardLocator(ctx, database.runtime, selector, iamGuardTenantA, 1,
		"expired", now.Add(-72*time.Hour), now.Add(-71*time.Hour),
		now.Add(-26*time.Hour), now.Add(-time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	err = database.runtime.QueryRowContext(ctx, `SELECT * FROM
platform.iam_resolve_refresh_target_v2($1)`, selector).
		Scan(new(string), new(string), new(string), new([]byte), new(int64),
			new([]byte), new(int64), new(int64), new(int64))
	assertIAMSQLState(t, err, "P5732")
}

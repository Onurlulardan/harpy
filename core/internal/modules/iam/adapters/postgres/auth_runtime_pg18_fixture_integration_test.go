//go:build integration

package postgres

import (
	"bytes"
	"context"
	"database/sql"
	"testing"
)

const (
	iamAuthIdentifierID = "61000000-0000-4000-8000-000000000001"
	iamAuthSessionID    = "62000000-0000-4000-8000-000000000001"
	iamAuthFamilyID     = "63000000-0000-4000-8000-000000000001"
	iamAuthTokenID      = "64000000-0000-4000-8000-000000000001"
	iamExpiredSessionID = "62000000-0000-4000-8000-000000000002"
	iamExpiredFamilyID  = "63000000-0000-4000-8000-000000000002"
	iamExpiredTokenID   = "64000000-0000-4000-8000-000000000002"
)

type iamAuthRuntimeFixture struct {
	tenant        iamFoundationTenant
	lookupHMAC    []byte
	sourceSHA256  []byte
	credentialTag []byte
	selector      []byte
	secretHMAC    []byte
	locatorSHA    []byte
	expiredSelect []byte
	expiredSecret []byte
	expiredSHA    []byte
}

func seedIAMAuthRuntimeFixtures(
	t *testing.T,
	ctx context.Context,
	database iamPlatformGuardDatabase,
) []iamAuthRuntimeFixture {
	t.Helper()
	fixtures := make([]iamAuthRuntimeFixture, 0, len(iamFoundationTenants))
	for index, tenant := range iamFoundationTenants {
		base := byte(0x71 + index*8)
		fixture := iamAuthRuntimeFixture{
			tenant: tenant, lookupHMAC: bytes.Repeat([]byte{base}, 32),
			sourceSHA256:  bytes.Repeat([]byte{base + 1}, 32),
			credentialTag: bytes.Repeat([]byte{base + 2}, 32),
			selector:      bytes.Repeat([]byte{base + 3}, 16),
			secretHMAC:    bytes.Repeat([]byte{base + 4}, 32),
			expiredSelect: bytes.Repeat([]byte{base + 5}, 16),
			expiredSecret: bytes.Repeat([]byte{base + 6}, 32),
		}
		fixture.locatorSHA = iamAuthLocatorSourceSHA256(t, iamAuthTokenID,
			iamAuthFamilyID, fixture.selector, 1, "active", fixture.secretHMAC)
		fixture.expiredSHA = iamAuthLocatorSourceSHA256(t, iamExpiredTokenID,
			iamExpiredFamilyID, fixture.expiredSelect, 1, "active",
			fixture.expiredSecret)
		mustExecIAMTenant(t, ctx, database.migrator, tenant.schema,
			seedIAMAuthUserStateSQL)
		mustExecIAMTenant(t, ctx, database.migrator, tenant.schema,
			seedIAMAuthIdentifierSQL, fixture.sourceSHA256)
		mustExecIAMTenant(t, ctx, database.migrator, tenant.schema,
			seedIAMAuthPasswordSQL, fixture.credentialTag)
		mustExecIAMTenant(t, ctx, database.migrator, tenant.schema,
			seedIAMExpiredSessionSQL)
		mustExecIAMTenant(t, ctx, database.migrator, tenant.schema,
			seedIAMExpiredFamilySQL)
		mustExecIAMTenant(t, ctx, database.migrator, tenant.schema,
			seedIAMExpiredTokenSQL, fixture.expiredSelect, fixture.expiredSecret)
		_, err := database.migrator.ExecContext(ctx, seedIAMLoginDirectorySQL,
			tenant.id, fixture.lookupHMAC, fixture.sourceSHA256)
		if err != nil {
			t.Fatalf("seed %s IAM login directory: %v", tenant.schema, err)
		}
		fixtures = append(fixtures, fixture)
	}
	return fixtures
}

const seedIAMLoginDirectorySQL = `
INSERT INTO platform.iam_login_directory (
 tenant_id,identifier_kind,lookup_key_id,identifier_hmac,user_id,
 source_identifier_id,source_revision,state,source_sha256,projected_at
) VALUES ($1,'email_v1','login-directory-v1',$2,
 '51000000-0000-4000-8000-000000000001',
 '61000000-0000-4000-8000-000000000001',1,'active',$3,
 pg_catalog.transaction_timestamp())`

func queryIAMAuthState(
	t *testing.T, ctx context.Context, database *sql.DB, schema, query string,
	destinations ...any,
) {
	t.Helper()
	mustQueryIAMTenantValue(t, ctx, database, schema, query, destinations...)
}

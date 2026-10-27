//go:build integration

package postgres

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

var iamInitialPasswordMigrationNames = []string{
	"20260716014500_create_iam_lifecycle_audit",
	"20260716014501_create_iam_lifecycle_receipt_v1",
	"20260722091000_create_iam_installation_password_credentials",
	"20260722091003_create_iam_initial_password_continuations",
	"20260722091004_create_iam_initial_password_credential_v1",
	"20260722091005_create_iam_initial_password_start_v1",
	"20260722091006_create_iam_initial_password_continuation_v1",
	"20260722091007_create_iam_initial_password_reject_v1",
	"20260722091008_create_iam_initial_password_complete_v1",
}

type iamInitialPasswordFixture struct {
	tenant                         iamFoundationTenant
	operationID, credentialID      uuid.UUID
	selector, secretHMAC, csrfHMAC []byte
}

func applyIAMInitialPassword(
	t *testing.T, ctx context.Context, database iamPlatformGuardDatabase, origin string,
) {
	t.Helper()
	documents := make([]iamFoundationDocument, 0, len(iamInitialPasswordMigrationNames))
	for _, name := range iamInitialPasswordMigrationNames {
		documents = append(documents, iamFoundationDocument{
			path: name, content: readIAMIdentityMigration(t, origin, name),
		})
	}
	for _, tenant := range iamFoundationTenants {
		applyIAMTenantDocuments(t, ctx, database.migrator, tenant.schema, documents)
	}
}

func seedIAMInitialPasswordFixtures(
	t *testing.T, ctx context.Context, database iamPlatformGuardDatabase,
) []iamInitialPasswordFixture {
	t.Helper()
	fixtures := make([]iamInitialPasswordFixture, 0, len(iamFoundationTenants))
	for index, tenant := range iamFoundationTenants {
		operationID, err := uuid.NewV7()
		if err != nil {
			t.Fatal(err)
		}
		fixture := iamInitialPasswordFixture{
			tenant: tenant, operationID: operationID, credentialID: uuid.New(),
			selector:   bytes.Repeat([]byte{byte(0x31 + index)}, 16),
			secretHMAC: bytes.Repeat([]byte{byte(0x41 + index)}, 32),
			csrfHMAC:   bytes.Repeat([]byte{byte(0x51 + index)}, 32),
		}
		mustExecIAMTenant(t, ctx, database.migrator, tenant.schema, `
INSERT INTO iam_password_credentials (
 user_id,credential_version,state,algorithm,salt,tag,memory_kib,iterations,
 parallelism,legacy_bcrypt_hash,upgrade_required,revision,created_at,updated_at
) VALUES ($1,1,'reset_required','argon2id_v1',$2,$3,65536,3,4,NULL,false,1,
 pg_catalog.transaction_timestamp(),pg_catalog.transaction_timestamp())`,
			iamUserOneID, bytes.Repeat([]byte{0x61}, 16),
			bytes.Repeat([]byte{byte(0x71 + index)}, 32))
		mustExecIAMTenant(t, ctx, database.migrator, tenant.schema, `
INSERT INTO iam_installation_password_credentials (
 credential_id,user_id,installation_operation_id,mutation_operation_id,
 credential_version,purpose,state,revision,created_at,updated_at
) VALUES ($2,$1,$3,$3,1,'installation_initial_v1','active',1,
 pg_catalog.transaction_timestamp(),pg_catalog.transaction_timestamp())`,
			iamUserOneID, fixture.credentialID, fixture.operationID)
		fixtures = append(fixtures, fixture)
	}
	return fixtures
}

func claimIAMInitialPassword(
	t *testing.T, ctx context.Context, database iamPlatformGuardDatabase,
	fixture iamInitialPasswordFixture,
) int64 {
	t.Helper()
	transaction, err := beginIAMAuthRuntimeTx(ctx, database.runtime, fixture.tenant.schema)
	if err != nil {
		t.Fatal(err)
	}
	var revision int64
	var issuedAt, expiresAt time.Time
	err = transaction.QueryRowContext(ctx, `SELECT * FROM iam_initial_password_start_v1(
 (pg_catalog.current_schemas(false))[2]::name,$1,$2,$3,$4,1,1,$5,$6,$7,$8,$9)`,
		fixture.tenant.id, iamUserOneID, fixture.operationID, fixture.credentialID,
		fixture.selector, "ABCDEFGHIJKLMNOP", fixture.secretHMAC,
		"QRSTUVWXYZabcdef", fixture.csrfHMAC,
	).Scan(&revision, &issuedAt, &expiresAt)
	if err != nil {
		_ = transaction.Rollback()
		t.Fatal(err)
	}
	if err := transaction.Commit(); err != nil {
		t.Fatal(err)
	}
	return revision
}

//go:build integration

package postgres

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

const (
	iamIdentityActorID      = "51000000-0000-4000-8000-000000000001"
	iamIdentityPermissionID = "93000000-0000-4000-8000-000000000002"
	iamIdentityOrgID        = "a2000000-0000-4000-8000-000000000001"
	iamIdentityOrgSchemaID  = "41000000-0000-4000-8000-000000000002"
)

const iamIdentityOrganizationCreate = `
SELECT result_id,result_revision,authorization_clock,replayed
FROM iam_identity_organization_create_v3(
    (pg_catalog.current_schemas(false))[2]::name,
    $1,$2,$3,$4,$5,NULL,
    ARRAY['51000000-0000-4000-8000-000000000003'::uuid],ARRAY[$3::uuid],
    $6,$7,$8,$9,$10,$11,$12)`

const iamIdentityOrganizationUpdate = `
SELECT result_id,result_revision,authorization_clock,replayed
FROM iam_identity_organization_update_v3(
    (pg_catalog.current_schemas(false))[2]::name,
    $1,$2,$3,$4,$5,$6,NULL,false,ARRAY[]::uuid[],ARRAY[]::uuid[],
    $7,$8,$9,$10,$11,$12,$13)`

func assertIAMIdentityReplayAndCAS(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	actor := readIAMIdentityActor(t, ctx, database, "t_iam_a", iamIdentityActorID)
	profile, profileDigest := []byte(`{}`), sha256.Sum256([]byte(`{}`))
	request := bytes.Repeat([]byte{0x31}, 32)
	withIAMIdentityRuntimeTx(t, ctx, database, "t_iam_a", func(transaction *sql.Tx) {
		for attempt := 0; attempt < 2; attempt++ {
			var id string
			var revision, clock int64
			var replayed bool
			err := transaction.QueryRowContext(ctx, iamIdentityOrganizationCreate,
				"a3000000-0000-4000-8000-000000000001", request,
				iamIdentityOrgID, "Runtime Organization", "runtime-organization",
				iamIdentityOrgSchemaID, profile, profileDigest[:], iamIdentityActorID,
				iamIdentityPermissionID, actor.tokenVersion, actor.authorizationVersion,
			).Scan(&id, &revision, &clock, &replayed)
			if err != nil || id != iamIdentityOrgID || revision != 1 || clock < 1 || replayed != (attempt == 1) {
				t.Fatalf("organization create attempt %d = %s/%d/%d/%t: %v",
					attempt, id, revision, clock, replayed, err)
			}
		}
		var id string
		var revision, clock int64
		var replayed bool
		err := transaction.QueryRowContext(ctx, iamIdentityOrganizationUpdate,
			"a3000000-0000-4000-8000-000000000002", bytes.Repeat([]byte{0x32}, 32),
			iamIdentityOrgID, int64(1), "Runtime Organization V2", "runtime-organization-v2",
			iamIdentityOrgSchemaID, profile, profileDigest[:], iamIdentityActorID,
			iamIdentityPermissionID, actor.tokenVersion, actor.authorizationVersion,
		).Scan(&id, &revision, &clock, &replayed)
		if err != nil || revision != 2 || replayed {
			t.Fatalf("organization CAS update = %s/%d/%d/%t: %v", id, revision, clock, replayed, err)
		}
	})
	probeIAMIdentityRuntimeTx(t, ctx, database, "t_iam_a", func(transaction *sql.Tx) {
		row := transaction.QueryRowContext(ctx, iamIdentityOrganizationUpdate,
			"a3000000-0000-4000-8000-000000000003", bytes.Repeat([]byte{0x33}, 32),
			iamIdentityOrgID, int64(1), "Stale", "stale-organization",
			iamIdentityOrgSchemaID, profile, profileDigest[:], iamIdentityActorID,
			iamIdentityPermissionID, actor.tokenVersion, actor.authorizationVersion)
		var id string
		var revision, clock int64
		var replayed bool
		assertIAMIdentitySQLState(t, row.Scan(&id, &revision, &clock, &replayed), "40001")
	})
}

func assertIAMIdentitySQLState(t *testing.T, err error, state string) {
	t.Helper()
	var postgresError *pgconn.PgError
	if !errors.As(err, &postgresError) || postgresError.Code != state {
		t.Fatalf("IAM identity SQLSTATE = %v, want %s", err, state)
	}
}

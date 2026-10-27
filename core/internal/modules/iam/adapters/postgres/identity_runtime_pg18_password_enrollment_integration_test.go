//go:build integration

package postgres

import (
	"bytes"
	"context"
	"database/sql"
	"testing"
)

const iamIdentityPasswordEnroll = `
SELECT result_id,result_revision,authorization_clock,replayed
FROM iam_identity_password_enroll_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1,$2,$3,$4,$5,$6,$7,$8,$9)`

func assertIAMIdentityPasswordEnrollment(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	for index, tenant := range iamFoundationTenants {
		actor := readIAMIdentityActor(t, ctx, database, tenant.schema, iamIdentityActorID)
		salt := bytes.Repeat([]byte{byte(0x41 + index)}, 16)
		tag := bytes.Repeat([]byte{byte(0x51 + index)}, 32)
		request := bytes.Repeat([]byte{byte(0x61 + index)}, 32)
		operation := "a7000000-0000-4000-8000-00000000000" + string(rune('1'+index))
		withIAMIdentityRuntimeTx(t, ctx, database, tenant.schema, func(transaction *sql.Tx) {
			for attempt := 0; attempt < 2; attempt++ {
				var id string
				var revision, clock int64
				var replayed bool
				err := transaction.QueryRowContext(
					ctx, iamIdentityPasswordEnroll, operation, request,
					iamIdentityActorID, salt, tag, iamIdentityActorID,
					iamIdentityPermissionID, actor.tokenVersion, actor.authorizationVersion,
				).Scan(&id, &revision, &clock, &replayed)
				if err != nil || id != iamIdentityActorID || revision != 1 ||
					clock < 1 || replayed != (attempt == 1) {
					t.Fatalf("%s password enrollment attempt %d = %s/%d/%d/%t: %v",
						tenant.schema, attempt, id, revision, clock, replayed, err)
				}
			}
		})
		assertIAMPasswordCredential(t, ctx, database, tenant.schema, salt, tag)
	}
	assertIAMPasswordEnrollmentRollback(t, ctx, database)
}

func assertIAMPasswordCredential(
	t *testing.T, ctx context.Context, database *sql.DB,
	schema string, salt, tag []byte,
) {
	t.Helper()
	connection, transaction := beginIAMTenantTx(t, ctx, database, schema)
	defer connection.Close()
	defer transaction.Rollback()
	var algorithm, state string
	var storedSalt, storedTag []byte
	var memory, iterations int
	var parallelism int16
	err := transaction.QueryRowContext(ctx, `
SELECT algorithm,state,salt,tag,memory_kib,iterations,parallelism
FROM iam_password_credentials WHERE user_id=$1`, iamIdentityActorID).Scan(
		&algorithm, &state, &storedSalt, &storedTag, &memory, &iterations, &parallelism,
	)
	if err != nil || algorithm != "argon2id_v1" || state != "active" ||
		!bytes.Equal(storedSalt, salt) || !bytes.Equal(storedTag, tag) ||
		memory != 65536 || iterations != 3 || parallelism != 4 {
		t.Fatalf("%s stored password profile is invalid: %v", schema, err)
	}
}

func assertIAMPasswordEnrollmentRollback(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	actor := readIAMIdentityActor(t, ctx, database, "t_iam_a", iamIdentityActorID)
	probeIAMIdentityRuntimeTx(t, ctx, database, "t_iam_a", func(transaction *sql.Tx) {
		row := transaction.QueryRowContext(
			ctx, iamIdentityPasswordEnroll,
			"a7000000-0000-4000-8000-000000000003", bytes.Repeat([]byte{0x71}, 32),
			"51000000-0000-4000-8000-000000000002", bytes.Repeat([]byte{0x72}, 16),
			bytes.Repeat([]byte{0x73}, 32), iamIdentityActorID,
			iamIdentityPermissionID, actor.tokenVersion, actor.authorizationVersion,
		)
		var id string
		var revision, clock int64
		var replayed bool
		if err := row.Scan(&id, &revision, &clock, &replayed); err != nil {
			t.Fatal(err)
		}
	})
	connection, transaction := beginIAMTenantTx(t, ctx, database, "t_iam_a")
	defer connection.Close()
	defer transaction.Rollback()
	var count int
	if err := transaction.QueryRowContext(ctx, `
SELECT pg_catalog.count(*) FROM iam_password_credentials
WHERE user_id='51000000-0000-4000-8000-000000000002'`).Scan(&count); err != nil || count != 0 {
		t.Fatalf("rolled-back credential count=%d error=%v", count, err)
	}
}

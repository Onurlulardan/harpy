//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"errors"
	"testing"
)

func assertIAMIdentityTenantIsolation(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	actor := readIAMIdentityActor(t, ctx, database, "t_iam_b", iamIdentityActorID)
	withIAMIdentityRuntimeTx(t, ctx, database, "t_iam_b", func(transaction *sql.Tx) {
		var id string
		err := transaction.QueryRowContext(ctx, `
SELECT id FROM iam_identity_organization_read_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1,$2,$3,$4,$5)`, iamIdentityOrgID, iamIdentityActorID,
			iamIdentityPermissionID, actor.tokenVersion,
			actor.authorizationVersion).Scan(&id)
		if !errors.Is(err, sql.ErrNoRows) {
			t.Fatalf("tenant B observed tenant A organization: id=%s error=%v", id, err)
		}
	})
}

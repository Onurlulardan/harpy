//go:build integration

package postgres

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"testing"

	"github.com/google/uuid"
)

func assertIAMSelfProfileRollback(
	t *testing.T, ctx context.Context, database *sql.DB, actorID uuid.UUID,
) {
	t.Helper()
	actor := readIAMIdentityActor(t, ctx, database, "t_iam_a", iamUserThreeID)
	connection, transaction, err := beginIAMProfileCommandRuntimeTx(
		ctx, database, "t_iam_a",
	)
	if err != nil {
		t.Fatal(err)
	}
	operationID := uuid.MustParse("e2000000-0000-4000-8000-000000000001")
	request := sha256.Sum256([]byte("self-profile-rollback"))
	document, digest := []byte(`{}`), sha256.Sum256([]byte(`{}`))
	var result iamProfileCommandResult
	err = transaction.QueryRowContext(
		ctx, callIAMSelfProfileUpdate,
		operationID, request[:], actorID,
		uuid.MustParse(iamUserProfileSchemaID), uuid.MustParse(iamUserProfileSchemaID),
		document, digest[:], actor.revision,
		actor.tokenVersion, actor.authorizationVersion,
	).Scan(&result.id, &result.revision, &result.clock, &result.replayed)
	if err != nil {
		_ = transaction.Rollback()
		_ = connection.Close()
		t.Fatal(err)
	}
	_ = transaction.Rollback()
	_ = connection.Close()
	after := readIAMIdentityActor(t, ctx, database, "t_iam_a", iamUserThreeID)
	var receipts int
	mustQueryIAMTenantValue(t, ctx, database, "t_iam_a", `
SELECT pg_catalog.count(*) FROM iam_authorization_operation_receipts
WHERE operation_id='e2000000-0000-4000-8000-000000000001'`, &receipts)
	if after != actor || receipts != 0 {
		t.Fatalf("self rollback actor=%+v want=%+v receipts=%d", after, actor, receipts)
	}
}

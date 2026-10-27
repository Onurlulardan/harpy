//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

func assertIAMInvitationDeliveryIsolation(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	tenant iamFoundationTenant,
	fixture iamInvitationRegistrationFixture,
) {
	t.Helper()
	connection, transaction := beginIAMTenantTx(t, ctx, database, tenant.schema)
	defer connection.Close()
	var state string
	if err := transaction.QueryRowContext(ctx, `SELECT pg_catalog.concat_ws('|',
 (SELECT pg_catalog.count(*) FROM iam_invitation_delivery_artifacts WHERE id=$1),
 (SELECT pg_catalog.count(*) FROM iam_invitation_delivery_worker_claims),
	(SELECT pg_catalog.count(*) FROM iam_invitation_delivery_worker_receipts))`, fixture.artifactID).Scan(&state); err != nil {
		_ = transaction.Rollback()
		t.Fatalf("query isolated delivery state: %v", err)
	}
	if err := transaction.Commit(); err != nil {
		t.Fatalf("commit isolated delivery query: %v", err)
	}
	if state != "1|0|0" {
		t.Fatalf("isolated tenant delivery state=%s", state)
	}
}

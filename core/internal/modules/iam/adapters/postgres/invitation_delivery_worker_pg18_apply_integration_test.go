//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

var iamInvitationDeliveryWorkerMigrations = []string{
	"20260716015500_create_iam_invitation_delivery_worker_claims",
	"20260716015501_create_iam_invitation_delivery_worker_receipts",
	"20260716015502_create_iam_invitation_delivery_worker_claim_v1",
	"20260716015503_create_iam_invitation_delivery_worker_finalize_v1",
}

func applyIAMInvitationDeliveryWorker(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	origin string,
) {
	t.Helper()
	documents := make([]iamFoundationDocument, 0, len(iamInvitationDeliveryWorkerMigrations))
	for _, name := range iamInvitationDeliveryWorkerMigrations {
		documents = append(documents, iamFoundationDocument{
			path:    name + "." + origin + ".sql",
			content: readIAMIdentityMigration(t, origin, name),
		})
	}
	for _, tenant := range iamFoundationTenants {
		applyIAMTenantDocuments(t, ctx, database, tenant.schema, documents)
	}
}

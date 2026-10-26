//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

func applyIAMSelfProfile(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	origin string,
) {
	t.Helper()
	documents := make([]iamFoundationDocument, 0, len(iamSelfProfileMigrationNames))
	for _, name := range iamSelfProfileMigrationNames {
		documents = append(documents, iamFoundationDocument{
			path:    name + "." + origin + ".sql",
			content: readIAMIdentityMigration(t, origin, name),
		})
	}
	for _, tenant := range iamFoundationTenants {
		applyIAMTenantDocuments(t, ctx, database, tenant.schema, documents)
	}
}

//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"
)

func TestSystemResourceLayoutAgainstPostgreSQL184(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 7*time.Minute)
	defer cancel()
	database := startLayoutPG18(t, ctx, "fresh")
	prepareLayoutPG18(t, ctx, database)
	assertLayoutInstallationAuthorityContractPG18(t, ctx, database)
	applyLayoutMigrations(t, ctx, database)
	t.Run("crud_and_publication", func(t *testing.T) {
		assertLayoutCRUDAndPublication(t, ctx, database)
	})
	t.Run("write_idempotency", func(t *testing.T) {
		assertLayoutWriteIdempotency(t, ctx, database)
	})
	t.Run("malformed_contracts", func(t *testing.T) {
		assertMalformedLayoutContracts(t, ctx, database)
	})
	t.Run("field_reference_evidence", func(t *testing.T) {
		assertLayoutReferenceEvidence(t, ctx, database)
	})
	t.Run("authority_and_isolation", func(t *testing.T) {
		assertLayoutAuthorityAndIsolation(t, ctx, database)
	})
	t.Run("rollback", func(t *testing.T) {
		assertLayoutRollback(t, ctx, database)
	})
	t.Run("version_concurrency", func(t *testing.T) {
		assertLayoutVersionConcurrency(t, ctx, database)
	})
}

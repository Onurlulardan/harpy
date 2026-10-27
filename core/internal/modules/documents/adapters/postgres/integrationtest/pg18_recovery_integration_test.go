//go:build integration

package integrationtest

import (
	"context"
	"testing"
	"time"
)

func TestDocumentsRecoveryAgainstPostgreSQL184(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 7*time.Minute)
	defer cancel()
	database := startDocumentsPG18(t, ctx)
	prepareDocumentsPG18(t, ctx, database)
	applyDocumentsMigrations(t, ctx, database)

	fixtures := []documentsRecoveryFixture{
		newDocumentsRecoveryFixture("documents_fresh_a",
			"018f0000-0000-7000-8000-000000000a01"),
		newDocumentsRecoveryFixture("documents_adopted_a",
			"018f0000-0000-7000-8000-000000000a02"),
	}
	for _, fixture := range fixtures {
		prepareDocumentsRecoveryIdentity(t, ctx, database, fixture)
		seedDocumentsRecoveryVersion(t, ctx, database, fixture)
		if fixture.schema == "documents_fresh_a" {
			seedDocumentsRecoveryTransients(t, ctx, database, fixture)
		}
		assertDocumentsRecoveryOwnerDenied(t, ctx, database, fixture)
		beginDocumentsRecoveryHold(t, ctx, database, fixture)
		assertDocumentsRecoveryDeletionFence(t, ctx, database, fixture)
		copyAndSealDocumentsRecovery(t, ctx, database, fixture)
		applyDocumentsRecoveryLocator(t, ctx, database, fixture)
		if fixture.schema == "documents_fresh_a" {
			terminalizeDocumentsRecoveryTransients(t, ctx, database, fixture)
		}
		releaseDocumentsRecoveryHold(t, ctx, database, fixture)
		assertDocumentsRecoveryResult(t, ctx, database, fixture)
	}
}

//go:build integration

package integrationtest

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestProfileRecoveryStorageAgainstPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh", "v1_adopted"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 7*time.Minute)
			defer cancel()
			db := startRecoveryStoragePG18(t, ctx, origin)
			assertProfileRecoveryACL(t, ctx, db)
			assertRecoveryRuntimeRoleBoundary(t, ctx, db)
			seedPublicationStartup(t, ctx, db, publicationMigrationDigest(t))
			seedRecoveryStorageSecondUser(t, ctx, db)
			runtime := openPublicationRuntime(t, ctx, db)
			setPublicationJobInsert(t, ctx, db, true)
			planned, err := runtime.application.PublishSchema(
				ctx, publicationClaims(t, db.tenant), publicationBackfillInput(),
			)
			if err != nil || planned.BackfillPlanID == nil {
				t.Fatalf("create recovery parent = %#v/%v", planned, err)
			}
			parentID := *planned.BackfillPlanID
			assertPublicationHistoricalNormalWrite(t, ctx, db, runtime, parentID)
			fenceRecoveryStoragePlan(t, ctx, db, parentID)
			assertRecoveryMixedInventory(t, ctx, db, parentID)
			assertRecoveryFencedWriteContracts(t, ctx, db)
			assertRecoveryLineageConstraints(t, ctx, db, parentID)
		})
	}
}

func startRecoveryStoragePG18(
	t *testing.T, ctx context.Context, origin string,
) publicationDatabase {
	t.Helper()
	db := startPublicationPG18(t, ctx, origin)
	preparePublicationCatalog(t, ctx, db)
	seedPublicationV1Snapshot(t, ctx, db)
	applyPublicationDocuments(t, ctx, db,
		publicationIAMDocuments(t, origin, "tenant", "foundation"), true)
	seedPublicationSource(t, ctx, db, publicationSourceSchema(t))
	seedPublicationCurrentSource(t, ctx, db)
	applyPublicationDocuments(t, ctx, db,
		publicationIAMDocuments(t, origin, "tenant", "runtime"), true)
	recoveryDocuments := recoveryStorageDocuments(t, origin)
	assertRecoveryStorageDocumentBounds(t, recoveryDocuments)
	seedPublicationAuthority(t, ctx, db)
	applyPublicationDocuments(t, ctx, db,
		publicationIAMDocuments(t, origin, "platform", "platform"), false)
	applyPublicationDocuments(t, ctx, db,
		publicationAsyncDocuments(t, origin), false)
	grantPublicationRuntimeAccess(t, ctx, db)
	return db
}

func recoveryStoragePlanID(t *testing.T, value string) uuid.UUID {
	t.Helper()
	id, err := uuid.Parse(value)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

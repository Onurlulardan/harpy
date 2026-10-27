//go:build integration

package integrationtest

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestProfileWorkerCommitWriterRaceAgainstPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh", "v1_adopted"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(
				context.Background(), 7*time.Minute,
			)
			defer cancel()
			db := startPublicationPG18(t, ctx, origin)
			preparePublicationCatalog(t, ctx, db)
			seedPublicationV1Snapshot(t, ctx, db)
			applyPublicationDocuments(t, ctx, db,
				publicationIAMDocuments(t, origin, "tenant", "foundation"), true)
			source := publicationSourceSchema(t)
			seedPublicationSource(t, ctx, db, source)
			seedPublicationCurrentSource(t, ctx, db)
			applyPublicationDocuments(t, ctx, db,
				publicationIAMDocuments(t, origin, "tenant", "runtime"), true)
			seedPublicationAuthority(t, ctx, db)
			seedPublicationSelfIdentifier(t, ctx, db)
			applyPublicationDocuments(t, ctx, db,
				publicationIAMDocuments(t, origin, "platform", "platform"), false)
			applyPublicationDocuments(t, ctx, db,
				publicationAsyncDocuments(t, origin), false)
			grantPublicationRuntimeAccess(t, ctx, db)
			seedPublicationStartup(
				t, ctx, db, publicationMigrationDigest(t),
			)
			preparePublicationRaceTenant(t, ctx, db)
			runtime := openPublicationRuntime(t, ctx, db)
			planID := createPublicationRacePlan(t, ctx, db, runtime)
			assertPublicationWorkerCommitWriterRace(
				t, ctx, db, runtime, planID,
			)
		})
	}
}

func createPublicationRacePlan(
	t *testing.T, ctx context.Context, db publicationDatabase,
	runtime publicationRuntime,
) uuid.UUID {
	t.Helper()
	setPublicationJobInsert(t, ctx, db, true)
	planned, err := runtime.application.PublishSchema(
		ctx, publicationClaims(t, db.tenant), publicationBackfillInput(),
	)
	if err != nil || planned.BackfillPlanID == nil ||
		planned.ID != uuid.MustParse(publicationTargetV3ID) ||
		planned.PublicationState != "planned" || planned.Replayed {
		t.Fatalf("worker commit race plan = %#v/%v", planned, err)
	}
	assertPublicationPlanAndFirstJob(
		t, ctx, db, *planned.BackfillPlanID,
	)
	return *planned.BackfillPlanID
}

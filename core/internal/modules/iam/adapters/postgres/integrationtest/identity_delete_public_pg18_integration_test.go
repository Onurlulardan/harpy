//go:build integration

package integrationtest

import (
	"context"
	"testing"
	"time"
)

func TestIAMSystemResourcePublicUserDeleteAgainstPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh", "v1_adopted"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
			defer cancel()

			db := startPublicationPG18(t, ctx, origin)
			preparePublicationCatalog(t, ctx, db)
			seedPublicationV1Snapshot(t, ctx, db)
			applyPublicationDocuments(t, ctx, db,
				publicationIAMDocuments(t, origin, "tenant", "foundation"), true)
			seedPublicationSource(t, ctx, db, publicationSourceSchema(t))
			seedPublicationCurrentSource(t, ctx, db)
			seedPublicationEmptyMappingSource(t, ctx, db)
			seedPublicationPublicDeleteGraph(t, ctx, db)
			applyPublicationDocuments(t, ctx, db,
				publicationIAMDocuments(t, origin, "tenant", "runtime"), true)
			seedPublicationAuthority(t, ctx, db)
			seedPublicationPublicDeleteAuthority(t, ctx, db)
			applyPublicationDocuments(t, ctx, db,
				publicationIAMDocuments(t, origin, "platform", "platform"), false)
			seedPublicationPublicDeleteDirectory(t, ctx, db)
			applyPublicationDocuments(t, ctx, db,
				publicationAsyncDocuments(t, origin), false)
			grantPublicationRuntimeAccess(t, ctx, db)
			seedPublicationStartup(t, ctx, db, publicationMigrationDigest(t))

			runtime := openPublicationRuntime(t, ctx, db)
			assertPublicationPublicUserDelete(t, ctx, db, runtime)
		})
	}
}

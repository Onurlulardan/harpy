//go:build integration

package integrationtest

import (
	"context"
	"testing"

	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
)

type recoveryPG18Scenario struct {
	db       publicationDatabase
	runtime  publicationRuntime
	userV2   profileapp.Schema
	userV3   profileapp.Schema
	userRows []recoveryPlanRow
}

func prepareRecoveryPG18Scenario(
	t *testing.T, ctx context.Context, origin string,
) recoveryPG18Scenario {
	t.Helper()
	db := startPublicationPG18(t, ctx, origin)
	preparePublicationCatalog(t, ctx, db)
	applyPublicationDocuments(t, ctx, db,
		publicationIAMDocuments(t, origin, "tenant", "foundation"), true)
	seedPublicationSource(t, ctx, db, publicationSourceSchema(t))
	seedPublicationCurrentSource(t, ctx, db)
	seedPublicationEmptyMappingSource(t, ctx, db)
	applyPublicationDocuments(t, ctx, db,
		publicationIAMDocuments(t, origin, "tenant", "runtime"), true)
	seedPublicationAuthority(t, ctx, db)
	seedPublicationEmptyMappingAuthority(t, ctx, db)
	seedPublicationSelfIdentifier(t, ctx, db)
	applyPublicationDocuments(t, ctx, db,
		publicationIAMDocuments(t, origin, "platform", "platform"), false)
	applyPublicationDocuments(t, ctx, db,
		publicationAsyncDocuments(t, origin), false)
	grantPublicationRuntimeAccess(t, ctx, db)
	assertProfileRecoveryACL(t, ctx, db)
	seedPublicationStartup(t, ctx, db, publicationMigrationDigest(t))
	setPublicationJobInsert(t, ctx, db, true)
	runtime := openPublicationRuntime(t, ctx, db)
	return recoveryPG18Scenario{
		db: db, runtime: runtime,
		userV2: publicationCurrentSchema(t), userV3: recoveryUserTargetV3(t),
		userRows: seedRecoveryUserRows(t, ctx, db),
	}
}

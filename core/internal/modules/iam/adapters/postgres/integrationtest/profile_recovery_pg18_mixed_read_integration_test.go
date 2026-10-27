//go:build integration

package integrationtest

import (
	"bytes"
	"context"
	"testing"

	profilepostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/identity/profile"
	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/systemresource"
	identitydomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/identity"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func assertRecoveryMixedRead(
	t *testing.T, ctx context.Context, db publicationDatabase,
	runtime publicationRuntime, sourceID, targetID uuid.UUID,
) {
	t.Helper()
	var sourceRows, targetRows int
	err := db.admin.QueryRow(ctx, `SELECT
count(*) FILTER (WHERE profile_schema_version_id=$1),
count(*) FILTER (WHERE profile_schema_version_id=$2)
FROM t_profile_publication.iam_users`, sourceID, targetID).
		Scan(&sourceRows, &targetRows)
	if err != nil || sourceRows != 1 || targetRows != 2 {
		t.Fatalf("mixed recovery storage = %d/%d/%v", sourceRows, targetRows, err)
	}
	profiles, err := profilepostgres.NewStore()
	if err != nil {
		t.Fatal(err)
	}
	tenantID := uuid.MustParse(db.tenant.id)
	generations, err := runtime.pool.ResolveSnapshotDomainGenerations(
		ctx, tenantID, []string{"iam", "system_resource_layout"},
	)
	if err != nil {
		t.Fatal(err)
	}
	request, err := queryapp.SnapshotScopeRequest(
		uuid.MustParse(publicationActorID), tenantID,
		map[string]int64{"iam": generations["iam"]},
	)
	if err != nil {
		t.Fatal(err)
	}
	loaded := make(map[uuid.UUID][]byte, 2)
	err = runtime.coordinator.InTenant(
		ctx, request, func(scoped context.Context, scope platformtx.TenantScope) error {
			for _, id := range []uuid.UUID{
				uuid.MustParse(publicationActorID), uuid.MustParse(recoveryUserThreeID),
			} {
				document, loadErr := profiles.Load(
					scoped, scope, identitydomain.ProfileUser,
					identitydomain.IdentityID(id),
				)
				if loadErr != nil {
					return loadErr
				}
				if uuid.UUID(document.SchemaID()) != sourceID {
					t.Fatalf("mixed recovery projected schema = %s", document.SchemaID())
				}
				loaded[id] = document.Values()
			}
			return nil
		},
	)
	if err != nil || !bytes.Contains(
		loaded[uuid.MustParse(publicationActorID)], []byte("a@b.co"),
	) || !bytes.Contains(
		loaded[uuid.MustParse(recoveryUserThreeID)], []byte("e@f.co"),
	) {
		t.Fatalf("mixed recovery public read = %#v/%v", loaded, err)
	}
}

package systemtransfercleanup

import (
	"context"
	"crypto/sha256"
	"testing"
	"time"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	app "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
	"github.com/google/uuid"
)

var (
	cleanupTenantID = uuid.MustParse("9c65f0c0-0000-4000-8000-000000000001")
	cleanupActorID  = uuid.MustParse("9c65f0c0-0000-4000-8000-000000000002")
	cleanupSourceID = uuid.MustParse("019bffff-ffff-7fff-8fff-000000000003")
	cleanupLeaseID  = uuid.MustParse("019bffff-ffff-7fff-8fff-000000000004")
	cleanupNow      = time.Date(2026, 7, 18, 12, 0, 0, 0, time.UTC)
)

type fixtureRow func(...any) error

func (row fixtureRow) Scan(destinations ...any) error { return row(destinations...) }

type fixtureDatabase struct {
	query func(string, ...any) row
}

func (database fixtureDatabase) QueryRow(_ context.Context,
	statement string, arguments ...any,
) row {
	return database.query(statement, arguments...)
}

func cleanupTarget() app.SourceCleanupTarget {
	return app.SourceCleanupTarget{
		TenantID: cleanupTenantID, ActorID: cleanupActorID,
		SourceID: cleanupSourceID, Resource: publicresource.Users,
		ObjectKey: "iam-system-import-sources/v1/" + cleanupSourceID.String() +
			"/1.fcart",
		ObjectVersion: "version-1", LeaseOwnerID: cleanupLeaseID,
		Attempt: 2, LeaseGeneration: 3,
	}
}

func cleanupReceipt() app.SourceDeleteResult {
	return app.SourceDeleteResult{
		ProviderRequestID:    "provider-delete-1",
		DeleteResponseSHA256: sha256.Sum256([]byte("delete")),
		AbsenceProbeSHA256:   sha256.Sum256([]byte("absent")),
		VerifiedAbsentAt:     cleanupNow,
	}
}

func boundStore(t *testing.T, database fixtureDatabase) *Store {
	t.Helper()
	store := &Store{database: database, notFound: func(error) bool { return false }}
	store.inTenant = func(ctx context.Context, tenantID uuid.UUID,
		execute func(context.Context, *tenantStore) error,
	) error {
		if tenantID != cleanupTenantID {
			t.Fatalf("bound tenant = %s", tenantID)
		}
		return execute(ctx, &tenantStore{database: database,
			currentSchema: func(context.Context) (string, error) {
				return "tenant_cleanup", nil
			}, notFound: store.notFound})
	}
	return store
}

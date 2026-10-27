package systemtransfercleanup

import (
	"context"
	"errors"
	"testing"
	"time"

	app "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
	"github.com/google/uuid"
)

func TestClaimScansAndBindsOneExactTenant(t *testing.T) {
	target := cleanupTarget()
	database := fixtureDatabase{query: func(statement string, arguments ...any) row {
		switch statement {
		case nextTenantSQL:
			if len(arguments) != 1 || arguments[0] != nil {
				t.Fatalf("next tenant arguments = %#v", arguments)
			}
			return fixtureRow(func(destinations ...any) error {
				*destinations[0].(*uuid.UUID) = cleanupTenantID
				*destinations[1].(*string) = "tenant_cleanup"
				return nil
			})
		case claimSQL:
			if len(arguments) != 4 || arguments[0] != "tenant_cleanup" ||
				arguments[1] != cleanupTenantID || arguments[2] != cleanupLeaseID ||
				arguments[3] != cleanupNow {
				t.Fatalf("claim arguments = %#v", arguments)
			}
			return fixtureRow(func(destinations ...any) error {
				*destinations[0].(*uuid.UUID) = target.SourceID
				*destinations[1].(*uuid.UUID) = target.ActorID
				*destinations[2].(*string) = string(target.Resource)
				*destinations[3].(*string) = target.ObjectKey
				*destinations[4].(*string) = target.ObjectVersion
				*destinations[5].(*int) = target.Attempt
				*destinations[6].(*int64) = target.LeaseGeneration
				return nil
			})
		default:
			t.Fatalf("unexpected statement: %s", statement)
			return nil
		}
	}}
	scan, err := boundStore(t, database).AcquireSourceCleanup(
		context.Background(), app.SourceCleanupClaim{
			LeaseOwnerID: cleanupLeaseID, ObservedAt: cleanupNow,
		},
	)
	if err != nil || scan.ScannedTenantID != cleanupTenantID || !scan.Found ||
		scan.Target != target {
		t.Fatalf("AcquireSourceCleanup() = %#v, %v", scan, err)
	}
}

func TestClaimRejectsNonUTCTime(t *testing.T) {
	store := &Store{database: fixtureDatabase{}, notFound: func(error) bool { return false },
		inTenant: func(context.Context, uuid.UUID,
			func(context.Context, *tenantStore) error) error {
			return nil
		}}
	_, err := store.AcquireSourceCleanup(context.Background(),
		app.SourceCleanupClaim{LeaseOwnerID: cleanupLeaseID,
			ObservedAt: cleanupNow.In(time.FixedZone("other", 3600))})
	if !errors.Is(err, app.ErrSourceCleanupInvalid) {
		t.Fatalf("error = %v", err)
	}
}

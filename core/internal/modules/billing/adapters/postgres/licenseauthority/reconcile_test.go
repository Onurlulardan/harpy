package licenseauthority

import (
	"context"
	"errors"
	"testing"
	"time"

	licenseapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/licenseauthority"
	"github.com/google/uuid"
)

func TestClaimRefreshScansStableNonceAndHead(t *testing.T) {
	operationID := uuid.MustParse("01900000-0000-7000-8000-000000000101")
	licenseID, organizationID := uuid.New(), uuid.New()
	environmentID := uuid.MustParse("01900000-0000-7000-8000-000000000102")
	now := time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)
	storage, err := NewStore(DatabaseAccess{QueryRow: func(
		context.Context, string, ...any,
	) Row {
		return rowFixture(func(values ...any) error {
			*values[0].(*uuid.UUID), *values[1].(*uuid.UUID) = operationID, uuid.New()
			*values[2].(*uuid.UUID), *values[3].(*uuid.UUID) = licenseID, organizationID
			*values[4].(*uuid.UUID), *values[5].(*string) = environmentID, "connected"
			*values[6].(*[]byte) = make([]byte, 32)
			*values[7].(*bool), *values[8].(*int64) = true, 4
			*values[9].(*string) = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
			*values[10].(*time.Time), *values[11].(*time.Time) = now, now.Add(time.Minute)
			return nil
		})
	}})
	if err != nil {
		t.Fatal(err)
	}
	claim, err := storage.AcquireRefreshLease(t.Context(), licenseapp.RefreshCommand{
		OperationID: operationID, Binding: licenseapp.Binding{
			LicenseID: licenseID, LicenseeOrganizationID: organizationID,
			EnvironmentID: environmentID, DeliveryKind: licenseapp.DeliveryConnected,
		},
	}, []licenseapp.ComponentVersion{{Component: "core-worker", Version: "v2.0.0"}})
	if err != nil || claim.Head == nil || claim.Head.Sequence != 4 || claim.Nonce != [32]byte{} {
		t.Fatalf("claim = %+v, %v", claim, err)
	}
}

func TestReconcileStoreRejectsCancelledContext(t *testing.T) {
	storage, err := NewStore(DatabaseAccess{QueryRow: func(
		context.Context, string, ...any,
	) Row {
		t.Fatal("database called")
		return nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err = storage.PersistVerifiedSnapshot(ctx, licenseapp.RefreshClaim{}, licenseapp.VerifiedSnapshot{})
	if !errors.Is(err, licenseapp.ErrAuthorityState) {
		t.Fatalf("error = %v", err)
	}
}

func TestRecordRefreshFailureAcceptsRetryableState(t *testing.T) {
	storage, err := NewStore(DatabaseAccess{QueryRow: func(
		context.Context, string, ...any,
	) Row {
		return rowFixture(func(values ...any) error {
			*values[0].(*string) = "pending"
			return nil
		})
	}})
	if err != nil {
		t.Fatal(err)
	}
	if err = storage.StoreRefreshFailure(
		t.Context(), licenseapp.RefreshClaim{}, licenseapp.FailureSource,
	); err != nil {
		t.Fatal(err)
	}
}

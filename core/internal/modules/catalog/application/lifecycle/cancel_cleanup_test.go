package lifecycle

import (
	"context"
	"testing"
	"time"

	catalogdomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/lifecycle"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/packagecontract"
	"github.com/google/uuid"
)

type cleanupClock struct{ value time.Time }

func (clock cleanupClock) Now() time.Time { return clock.value }

type cleanupStore struct {
	installation catalogdomain.Installation
	mutation     InstallationMutation
}

func (store *cleanupStore) CreateOperation(context.Context, StartMutation) (OperationHandle, error) {
	return OperationHandle{}, ErrUnavailable
}
func (store *cleanupStore) ReadOperation(context.Context, uuid.UUID) (catalogdomain.Operation, error) {
	return catalogdomain.Operation{}, ErrUnavailable
}
func (store *cleanupStore) ChangeOperation(context.Context, OperationMutation) (catalogdomain.Operation, error) {
	return catalogdomain.Operation{}, ErrUnavailable
}
func (store *cleanupStore) RecordOwnerReceipt(context.Context, catalogdomain.OwnerReceipt) error {
	return ErrUnavailable
}
func (store *cleanupStore) FinalizePurgeBlockerReceipt(context.Context, PurgeBlockerReceipt) error {
	return ErrUnavailable
}
func (store *cleanupStore) CompileOwnerReceipts(context.Context, uuid.UUID) ([]catalogdomain.OwnerReceipt, error) {
	return nil, ErrUnavailable
}
func (store *cleanupStore) GetInstallation(context.Context, packagecontract.PackageID) (catalogdomain.Installation, error) {
	return store.installation, nil
}
func (store *cleanupStore) UpdateInstallation(_ context.Context, mutation InstallationMutation) (catalogdomain.Installation, error) {
	if mutation.ExpectedRevision != store.installation.RowVersion {
		return catalogdomain.Installation{}, ErrConflict
	}
	store.mutation, store.installation = mutation, mutation.Installation
	return mutation.Installation, nil
}
func (store *cleanupStore) ListInstallations(context.Context, PageRequest) (InstallationPage, error) {
	return InstallationPage{}, ErrUnavailable
}
func (store *cleanupStore) ResolveDependencyBlockers(context.Context, packagecontract.PackageID) ([]DependencyBlocker, error) {
	return nil, ErrUnavailable
}

func TestCancelCleanupRestoresInstallationState(t *testing.T) {
	now := time.Date(2026, 7, 18, 13, 0, 0, 0, time.UTC)
	actorID := uuid.MustParse("018f4000-0000-7000-8000-000000000001")
	operationID := uuid.MustParse("018f4000-0000-7000-8000-000000000002")
	tests := []struct {
		name      string
		kind      catalogdomain.OperationKind
		input     catalogdomain.Installation
		wantState catalogdomain.InstallationState
		wantEpoch uint64
	}{
		{name: "install", kind: catalogdomain.OperationInstall, input: cleanupInstalling(t),
			wantState: catalogdomain.InstallationFailed, wantEpoch: 1},
		{name: "retry", kind: catalogdomain.OperationRetry, input: cleanupInstalling(t),
			wantState: catalogdomain.InstallationFailed, wantEpoch: 1},
		{name: "purge", kind: catalogdomain.OperationPurge, input: cleanupPurgePending(t),
			wantState: catalogdomain.InstallationRetired, wantEpoch: 4},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			operation := cleanupOperation(t, operationID, test.kind)
			store := &cleanupStore{installation: test.input}
			application := &Application{clock: cleanupClock{value: now}}
			err := application.cancelInstallation(context.Background(), store,
				ScopeIdentity{ActorID: actorID}, CancelRequest{OperationID: operationID,
					IdempotencyKey: "catalog-cancel-0001"}, operation)
			if err != nil {
				t.Fatal(err)
			}
			if store.installation.State != test.wantState || store.installation.ReferenceEpoch != test.wantEpoch ||
				store.mutation.ActorID != actorID || store.mutation.RequestSHA256 == "" {
				t.Fatalf("cleanup mutation = %#v", store.mutation)
			}
		})
	}
}

func cleanupInstalling(t *testing.T) catalogdomain.Installation {
	t.Helper()
	id, _ := packagecontract.ParseStableID("018f4000-0000-7000-8000-000000000003")
	packageID, _ := packagecontract.ParsePackageID("018f4000-0000-7000-8000-000000000004")
	value, err := catalogdomain.NewInstalling(id, packageID)
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func cleanupPurgePending(t *testing.T) catalogdomain.Installation {
	value := cleanupInstalling(t)
	version, _ := packagecontract.ParseVersion("1.0.0")
	value, _ = value.Activate(version)
	value, _ = value.Retire()
	value, _ = value.BeginPurge()
	return value
}

func cleanupOperation(t *testing.T, operationID uuid.UUID, kind catalogdomain.OperationKind) catalogdomain.Operation {
	t.Helper()
	packageID, _ := packagecontract.ParsePackageID("018f4000-0000-7000-8000-000000000004")
	componentID, _ := packagecontract.ParseComponentID("018f4000-0000-7000-8000-000000000005")
	version, _ := packagecontract.ParseVersion("1.0.0")
	plan, err := catalogdomain.NewPlan(packageID, version, nil, []catalogdomain.ComponentPin{{
		ComponentID: componentID, Owner: "catalog", OwnerGeneration: 1, Contract: "catalog.v2",
		ContentSHA256: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	}})
	if err != nil {
		t.Fatal(err)
	}
	stableID, _ := packagecontract.ParseStableID(operationID.String())
	value, err := catalogdomain.NewOperation(stableID, kind, plan,
		time.Date(2026, 7, 18, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	return value
}

package lifecycle

import (
	"context"
	"errors"
	"testing"

	catalogdomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/lifecycle"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func TestDemoRemoveRevalidatesPlanBeforeOwnerEffect(t *testing.T) {
	sealed, err := sealDemoRemovalPlan(demoInstallationID(), []DemoRemovalComponent{{
		ComponentID: demoComponentID(), SourceRunID: demoRunID(),
		RecordIDs: []uuid.UUID{demoRecordA()},
	}})
	if err != nil {
		t.Fatal(err)
	}
	store := &demoResumeStore{cleanupStore: &cleanupStore{},
		components: []DemoComponent{demoComponentFixture(
			DemoStateInstalled, []uuid.UUID{demoRecordA()},
		)},
		plan: DemoOperationPlan{OperationID: demoOperationID(),
			InstallationID: demoInstallationID(), Action: DemoActionRemove,
			ExpectedPlanSHA256: sealed.PlanSHA256,
			Components:         sealed.Components, CreatedAt: demoNow()},
	}
	owner := &demoOwnerFixture{planIDs: []uuid.UUID{demoRecordB()}}
	application := &Application{demo: owner}
	operation := demoApplyingOperation(t, catalogdomain.OperationDemoRemove)
	_, err = application.stageDemoApplying(
		context.Background(), platformtx.TenantScope{}, store, operation,
	)
	if !errors.Is(err, ErrDemoPlanStale) || owner.removeCalls != 0 ||
		store.applyCalls != 0 || store.changed.ID.Valid() {
		t.Fatalf("stale apply err=%v owner=%d store=%d changed=%#v",
			err, owner.removeCalls, store.applyCalls, store.changed)
	}
}

func TestDemoInstallPersistsReceiptBeforeAdvancing(t *testing.T) {
	store := &demoResumeStore{cleanupStore: &cleanupStore{},
		components: []DemoComponent{demoComponentFixture(DemoStateAbsent, nil)},
		plan: DemoOperationPlan{OperationID: demoOperationID(),
			InstallationID: demoInstallationID(), Action: DemoActionInstall,
			Components: []DemoRemovalComponent{{
				ComponentID: demoComponentID(), SourceRunID: demoRunID(),
			}}, CreatedAt: demoNow()},
	}
	owner := &demoOwnerFixture{installIDs: []uuid.UUID{demoRecordA()}}
	application := &Application{demo: owner}
	result, err := application.stageDemoApplying(context.Background(),
		platformtx.TenantScope{}, store,
		demoApplyingOperation(t, catalogdomain.OperationDemoInstall))
	if err != nil || result.Operation.State != catalogdomain.OperationVerifying ||
		owner.installCalls != 1 || store.applyCalls != 1 ||
		store.components[0].State != DemoStateInstalled ||
		store.changed.State != catalogdomain.OperationVerifying {
		t.Fatalf("install result=%#v err=%v owner=%d store=%#v",
			result, err, owner.installCalls, store)
	}
}

func TestDemoApplyingWithoutOwnerFailsClosed(t *testing.T) {
	store := &demoResumeStore{cleanupStore: &cleanupStore{}}
	_, err := (&Application{}).stageDemoApplying(context.Background(),
		platformtx.TenantScope{}, store,
		demoApplyingOperation(t, catalogdomain.OperationDemoInstall))
	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("nil owner error=%v", err)
	}
}

type wrongDemoSourceOwner struct{ *demoOwnerFixture }

func (owner *wrongDemoSourceOwner) Install(
	context.Context,
	platformtx.TenantScope,
	DemoOwnerCommand,
) (DemoOwnerResult, error) {
	owner.installCalls++
	return DemoOwnerResult{
		ReceiptID:   demoReceiptID(),
		SourceRunID: uuid.MustParse("018f5000-0000-7000-8000-000000000099"),
		Installed:   true, RecordIDs: []uuid.UUID{demoRecordA()},
		CompletedAt: demoNow(),
	}, nil
}

func TestDemoInstallRejectsOwnerSourceRunDrift(t *testing.T) {
	store := &demoResumeStore{cleanupStore: &cleanupStore{},
		components: []DemoComponent{demoComponentFixture(DemoStateAbsent, nil)},
		plan: DemoOperationPlan{OperationID: demoOperationID(),
			InstallationID: demoInstallationID(), Action: DemoActionInstall,
			Components: []DemoRemovalComponent{{
				ComponentID: demoComponentID(), SourceRunID: demoRunID(),
			}}, CreatedAt: demoNow()},
	}
	owner := &wrongDemoSourceOwner{demoOwnerFixture: &demoOwnerFixture{}}
	application := &Application{demo: owner}
	_, err := application.stageDemoApplying(context.Background(),
		platformtx.TenantScope{}, store,
		demoApplyingOperation(t, catalogdomain.OperationDemoInstall))
	if !errors.Is(err, ErrUnavailable) || store.applyCalls != 0 {
		t.Fatalf("source-run drift err=%v apply=%d", err, store.applyCalls)
	}
}

func demoApplyingOperation(
	t *testing.T,
	kind catalogdomain.OperationKind,
) catalogdomain.Operation {
	t.Helper()
	value := cleanupOperation(t, demoOperationID(), kind)
	value, err := value.Advance(catalogdomain.OperationApplying)
	if err != nil {
		t.Fatal(err)
	}
	return value
}

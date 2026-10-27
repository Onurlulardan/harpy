package lifecycle

import (
	"context"
	"time"

	catalogdomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/lifecycle"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type demoResumeStore struct {
	*cleanupStore
	plan       DemoOperationPlan
	components []DemoComponent
	changed    catalogdomain.Operation
	applyCalls int
}

func (store *demoResumeStore) ChangeOperation(
	_ context.Context,
	mutation OperationMutation,
) (catalogdomain.Operation, error) {
	store.changed = mutation.Operation
	return mutation.Operation, nil
}

func (store *demoResumeStore) DemoComponents(
	context.Context,
	uuid.UUID,
) ([]DemoComponent, error) {
	result := make([]DemoComponent, len(store.components))
	for index, component := range store.components {
		result[index] = component.Clone()
	}
	return result, nil
}

func (store *demoResumeStore) RecordDemoOperationPlan(
	context.Context,
	DemoOperationPlan,
) (DemoOperationPlan, error) {
	return DemoOperationPlan{}, ErrUnavailable
}

func (store *demoResumeStore) ReadDemoOperationPlan(
	context.Context,
	uuid.UUID,
) (DemoOperationPlan, error) {
	return store.plan, nil
}

func (store *demoResumeStore) ApplyDemoResult(
	_ context.Context,
	mutation DemoResultMutation,
) (DemoComponent, error) {
	store.applyCalls++
	for index := range store.components {
		component := &store.components[index]
		if component.ComponentID != mutation.ComponentID ||
			component.Revision != mutation.ExpectedRevision {
			continue
		}
		component.SourceRunID = mutation.SourceRunID
		component.Revision++
		component.UpdatedAt = mutation.CompletedAt
		component.RecordIDs = append([]uuid.UUID(nil), mutation.RecordIDs...)
		component.State = DemoStateInstalled
		if mutation.Action == DemoActionRemove {
			component.State, component.RecordIDs = DemoStateAbsent, nil
		}
		return component.Clone(), nil
	}
	return DemoComponent{}, ErrConflict
}

type demoOwnerFixture struct {
	planIDs      []uuid.UUID
	installIDs   []uuid.UUID
	removeCalls  int
	installCalls int
}

func (owner *demoOwnerFixture) Ready(context.Context, platformtx.TenantScope) error {
	return nil
}

func (owner *demoOwnerFixture) PlanRemoval(
	context.Context,
	platformtx.TenantScope,
	DemoOwnerCommand,
) (DemoOwnerResult, error) {
	return DemoOwnerResult{Installed: true, SourceRunID: demoRunID(),
		RecordIDs: append([]uuid.UUID(nil), owner.planIDs...)}, nil
}

func (owner *demoOwnerFixture) Install(
	context.Context,
	platformtx.TenantScope,
	DemoOwnerCommand,
) (DemoOwnerResult, error) {
	owner.installCalls++
	return DemoOwnerResult{ReceiptID: demoReceiptID(), SourceRunID: demoRunID(),
		Installed: true, RecordIDs: append([]uuid.UUID(nil), owner.installIDs...),
		CompletedAt: demoNow()}, nil
}

func (owner *demoOwnerFixture) Remove(
	_ context.Context,
	_ platformtx.TenantScope,
	command DemoOwnerCommand,
) (DemoOwnerResult, error) {
	owner.removeCalls++
	return DemoOwnerResult{ReceiptID: demoReceiptID(), SourceRunID: demoRunID(),
		RecordIDs:   append([]uuid.UUID(nil), command.RecordIDs...),
		CompletedAt: demoNow()}, nil
}

func (owner *demoOwnerFixture) Inspect(
	context.Context,
	platformtx.TenantScope,
	DemoOwnerCommand,
) (DemoOwnerResult, error) {
	return DemoOwnerResult{}, ErrUnavailable
}

func demoComponentFixture(state string, records []uuid.UUID) DemoComponent {
	return DemoComponent{
		InstallationID: demoInstallationID(),
		PackageID:      uuid.MustParse("76000000-0000-4000-8000-000000000011"),
		ComponentID:    demoComponentID(), ReleaseID: demoReleaseID(),
		SourceRunID: demoRunID(), PackageSlug: "sales",
		PackageVersion: "1.0.0", ContentSHA256: demoDigest(), State: state,
		Document: []byte(`{"demo":true}`), RecordIDs: records,
		Revision: 1, UpdatedAt: demoNow().Add(-time.Minute),
	}
}

func demoNow() time.Time            { return time.Date(2026, 8, 1, 10, 0, 0, 0, time.UTC) }
func demoInstallationID() uuid.UUID { return uuid.MustParse("018f5000-0000-7000-8000-000000000001") }
func demoOperationID() uuid.UUID    { return uuid.MustParse("018f5000-0000-7000-8000-000000000002") }
func demoComponentID() uuid.UUID    { return uuid.MustParse("018f5000-0000-7000-8000-000000000003") }
func demoReleaseID() uuid.UUID      { return uuid.MustParse("018f5000-0000-7000-8000-000000000004") }
func demoRunID() uuid.UUID          { return uuid.MustParse("018f5000-0000-7000-8000-000000000005") }
func demoReceiptID() uuid.UUID      { return uuid.MustParse("018f5000-0000-7000-8000-000000000006") }
func demoRecordA() uuid.UUID        { return uuid.MustParse("76000000-0000-4000-8000-000000000021") }
func demoRecordB() uuid.UUID        { return uuid.MustParse("76000000-0000-4000-8000-000000000022") }
func demoDigest() string            { return "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" }

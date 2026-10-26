package folder

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

var (
	fixtureTenant = fixtureID("11111111-1111-4111-8111-111111111111")
	fixtureActor  = fixtureID("22222222-2222-4222-8222-222222222222")
	fixtureFolder = fixtureID("018f0000-0000-7000-8000-000000000501")
	fixtureOp     = fixtureID("018f0000-0000-7000-8000-000000000502")
	fixtureAudit  = fixtureID("018f0000-0000-7000-8000-000000000503")
	fixtureNow    = time.Date(2026, 7, 18, 11, 0, 0, 0, time.UTC)
)

func fixtureID(value string) domain.ID {
	result, _ := domain.ParseID(value)
	return result
}

func fixtureFolderValue(version uint64) domain.Folder {
	value, _ := domain.NewFolder(fixtureFolder, fixtureTenant, domain.ID{},
		fixtureActor, "Contracts", fixtureNow)
	value.RowVersion = version
	return value
}

type fixtureStore struct {
	create  func(context.Context, Mutation) (domain.Folder, bool, error)
	get     func(context.Context, domain.ID) (domain.Folder, error)
	list    func(context.Context, domain.ID, domain.ID, uint16) (Page, error)
	update  func(context.Context, Mutation) (domain.Folder, bool, error)
	move    func(context.Context, Mutation) (domain.Folder, bool, error)
	delete  func(context.Context, Mutation) (domain.Folder, bool, error)
	restore func(context.Context, Mutation) (domain.Folder, bool, error)
}

func (store fixtureStore) Create(ctx context.Context, mutation Mutation) (domain.Folder, bool, error) {
	return store.create(ctx, mutation)
}
func (store fixtureStore) Get(ctx context.Context, id domain.ID) (domain.Folder, error) {
	return store.get(ctx, id)
}
func (store fixtureStore) List(ctx context.Context, parent, after domain.ID, limit uint16) (Page, error) {
	return store.list(ctx, parent, after, limit)
}
func (store fixtureStore) Update(ctx context.Context, mutation Mutation) (domain.Folder, bool, error) {
	return store.update(ctx, mutation)
}
func (store fixtureStore) Move(ctx context.Context, mutation Mutation) (domain.Folder, bool, error) {
	return store.move(ctx, mutation)
}
func (store fixtureStore) Delete(ctx context.Context, mutation Mutation) (domain.Folder, bool, error) {
	return store.delete(ctx, mutation)
}
func (store fixtureStore) Restore(ctx context.Context, mutation Mutation) (domain.Folder, bool, error) {
	return store.restore(ctx, mutation)
}

type fixtureIDs struct{ values []domain.ID }

func (source *fixtureIDs) NewUUIDv7() (domain.ID, error) {
	value := source.values[0]
	source.values = source.values[1:]
	return value, nil
}

func fixtureApplication() *Application {
	return &Application{ids: &fixtureIDs{values: []domain.ID{fixtureOp, fixtureAudit}}}
}

func fixtureIdentity() foundation.Identity {
	return foundation.Identity{ActorID: fixtureActor, TenantID: fixtureTenant}
}

func fixtureMutation(version uint64) Mutation {
	return Mutation{OperationID: fixtureOp, AuditID: fixtureAudit, ActorID: fixtureActor,
		FolderID: fixtureFolder, ExpectedVersion: version, IdempotencyKey: "folder-fixture-1",
		OccurredAt: fixtureNow}
}

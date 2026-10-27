package tasks

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

type updateResourceStore struct {
	ports.TaskStore
	task         domain.Task
	persisted    domain.Task
	currentLinks []domain.ResourceLink
	replacements []domain.ResourceLink
	replaced     bool
	legacy       domain.LegacyTaskMetadata
	legacyWrites int
	legacyErr    error
}

func (store *updateResourceStore) LoadTaskForUpdate(
	context.Context, domain.ID,
) (domain.Task, error) {
	return store.task, nil
}

func (store *updateResourceStore) PersistTask(
	_ context.Context, task domain.Task, _ int64,
) error {
	store.persisted = task
	return nil
}

func (store *updateResourceStore) LoadResourceLinksForUpdate(
	context.Context, domain.ID,
) ([]domain.ResourceLink, error) {
	return append([]domain.ResourceLink(nil), store.currentLinks...), nil
}

func (store *updateResourceStore) ReplaceResourceLinks(
	_ context.Context, _ domain.ID, replacements []domain.ResourceLink,
) error {
	store.replaced = true
	store.replacements = append([]domain.ResourceLink(nil), replacements...)
	return nil
}

func (store *updateResourceStore) UpsertLegacyTaskMetadata(
	_ context.Context, _ domain.ID, value domain.LegacyTaskMetadata,
) error {
	store.legacyWrites++
	store.legacy = value
	return store.legacyErr
}

type updateResourceStores struct {
	ports.Stores
	tasks *updateResourceStore
}

func (stores updateResourceStores) Tasks() ports.TaskStore { return stores.tasks }

type updateResourceMutator struct {
	stores ports.Stores
	hashes [][32]byte
	run    bool
}

func (mutator *updateResourceMutator) Mutate(
	ctx context.Context, _ ports.Scope, _ string, _ domain.ID, _ ports.Mutation,
	hash [32]byte, action ports.MutationAction,
) (ports.OperationReceipt, error) {
	mutator.hashes = append(mutator.hashes, hash)
	if !mutator.run {
		return ports.OperationReceipt{}, nil
	}
	receipt, _, err := action(ctx, mutator.stores)
	return receipt, err
}

type updateResourceAuthority struct {
	allowed  bool
	revision int64
	err      error
	calls    int
}

func (authority *updateResourceAuthority) AuthorizeRead(
	context.Context, ports.Scope, ports.ResourceAuthorityRequest,
) (bool, int64, error) {
	authority.calls++
	return authority.allowed, authority.revision, authority.err
}

type updatePrincipalResolver struct{ ports.PrincipalResolver }
type updateClock struct{}
type updateIDs struct{ next domain.ID }
type updateReferences struct {
	ports.ResourceReferenceRegistry
	registered []ports.ResourceReferenceCommand
	released   []ports.ResourceReferenceCommand
	err        error
}

func (references *updateReferences) RegisterRead(
	_ context.Context, _ ports.Scope, command ports.ResourceReferenceCommand,
) (ports.ResourceReferenceReceipt, error) {
	references.registered = append(references.registered, command)
	return referenceReceipt(command, 7), references.err
}
func (references *updateReferences) ReleaseRead(
	_ context.Context, _ ports.Scope, command ports.ResourceReferenceCommand,
) (ports.ResourceReferenceReceipt, error) {
	references.released = append(references.released, command)
	return referenceReceipt(command, command.Link.ReferenceEpoch), references.err
}

func referenceReceipt(command ports.ResourceReferenceCommand,
	epoch int64,
) ports.ResourceReferenceReceipt {
	return ports.ResourceReferenceReceipt{ReferenceID: command.Link.ID,
		OperationID: command.OperationID, TargetID: command.Link.TargetID,
		TargetVersionID: command.Link.TargetVersionID, ReferenceEpoch: epoch}
}

func (updateClock) Now() time.Time {
	return time.Date(2026, 8, 2, 12, 0, 0, 0, time.UTC)
}
func (ids *updateIDs) New(time.Time) (domain.ID, error) {
	result := ids.next
	ids.next[15]++
	return result, nil
}

func updateID(seed byte) domain.ID { var id domain.ID; id[15] = seed; return id }

func validUpdateTask(kind domain.TaskKind) domain.Task {
	return domain.Task{ID: updateID(1), BoardID: updateID(2), ColumnID: updateID(3),
		CreatorID: updateID(4), Kind: kind, Status: domain.StatusOpen,
		Priority: domain.PriorityMedium, Title: "Current", Position: 1024, RowVersion: 3}
}

func updateInput(task domain.Task) UpdateInput {
	return UpdateInput{TaskID: task.ID, Priority: task.Priority,
		Title: task.Title, Description: task.Description}
}

func updateResourceInput(seed byte) *ResourceLinkInput {
	return &ResourceLinkInput{Owner: domain.SourceForms, Targets: []ResourceTargetInput{
		{Kind: "FORM", ID: updateID(seed), VersionID: updateID(seed + 1)},
	}}
}

func storedUpdateResource(task domain.Task, seed byte) []domain.ResourceLink {
	return []domain.ResourceLink{{ID: updateID(seed), WorkResourceID: task.ID,
		TargetOwner: domain.SourceForms, TargetKind: "FORM",
		TargetID: updateID(seed + 1), TargetVersionID: updateID(seed + 2),
		ReferenceEpoch: 3, PermissionRevision: 4, RowVersion: 1}}
}

func updateService(
	mutator ports.Mutator, authority ports.ResourceAuthority,
) *Service {
	return updateServiceWithReferences(mutator, authority, &updateReferences{})
}

func updateServiceWithReferences(
	mutator ports.Mutator, authority ports.ResourceAuthority,
	references ports.ResourceReferenceRegistry,
) *Service {
	service, _ := New(Dependencies{Mutator: mutator,
		Principals: updatePrincipalResolver{}, Resources: authority,
		References: references, Clock: updateClock{},
		IDs: &updateIDs{next: updateID(90)}})
	return service
}

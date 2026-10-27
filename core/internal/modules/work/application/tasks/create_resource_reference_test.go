package tasks

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

type createBoardStore struct {
	ports.BoardStore
	board  domain.Board
	column domain.Column
}

func (store *createBoardStore) LoadBoardForUpdate(
	context.Context, domain.ID,
) (domain.Board, error) {
	return store.board, nil
}

func (store *createBoardStore) LoadColumnForUpdate(
	context.Context, domain.ID,
) (domain.Column, error) {
	return store.column, nil
}

func (*createBoardStore) AssigneesHaveAccess(
	context.Context, domain.ID, []domain.Principal,
) (bool, error) {
	return true, nil
}

type createTaskStore struct {
	ports.TaskStore
	task         domain.Task
	links        []domain.ResourceLink
	legacy       domain.LegacyTaskMetadata
	legacyWrites int
}

func (*createTaskStore) CountBoardTasksForUpdate(
	context.Context, domain.ID,
) (int, error) {
	return 0, nil
}

func (*createTaskStore) LoadRankedTasksForUpdate(
	context.Context, domain.ID,
) ([]domain.RankedTask, error) {
	return nil, nil
}

func (store *createTaskStore) InsertTask(_ context.Context, task domain.Task) error {
	store.task = task
	return nil
}

func (store *createTaskStore) InsertResourceLink(
	_ context.Context, link domain.ResourceLink,
) error {
	store.links = append(store.links, link)
	return nil
}

func (store *createTaskStore) UpsertLegacyTaskMetadata(
	_ context.Context, _ domain.ID, value domain.LegacyTaskMetadata,
) error {
	store.legacyWrites++
	store.legacy = value
	return nil
}

type createStores struct {
	ports.Stores
	boards *createBoardStore
	tasks  *createTaskStore
}

func (stores createStores) Boards() ports.BoardStore { return stores.boards }
func (stores createStores) Tasks() ports.TaskStore   { return stores.tasks }

type createScope struct{ actor, tenant domain.ID }

func (scope createScope) TenantID() domain.ID { return scope.tenant }
func (scope createScope) ActorID() domain.ID  { return scope.actor }
func (createScope) OperationID() string       { return "work.task.create" }
func (createScope) SnapshotQuery() bool       { return false }

type createPrincipalResolver struct{}

func (createPrincipalResolver) Resolve(
	_ context.Context, scope ports.Scope,
) (ports.PrincipalSnapshot, error) {
	return ports.PrincipalSnapshot{Principals: []domain.Principal{{
		Kind: domain.PrincipalUser, ID: scope.ActorID()}}, PermissionRevision: 11}, nil
}

func (createPrincipalResolver) Exists(
	context.Context, ports.Scope, domain.Principal,
) (bool, int64, error) {
	return true, 11, nil
}

func TestCreateFormTaskRegistersOwnerReferenceBeforeLinkPersistence(t *testing.T) {
	service, store, references := createResourceService(t, nil)
	resource := updateResourceInput(10)
	receipt, err := service.CreateTask(context.Background(), createScope{
		actor: updateID(8), tenant: updateID(9)}, ports.Mutation{
		IdempotencyKey: "create-form-task"}, CreateTaskInput{
		BoardID: updateID(2), ColumnID: updateID(3), Kind: domain.TaskFormFill,
		Priority: domain.PriorityMedium, Title: "Fill request", Resource: resource,
	})
	if err != nil || !receipt.AggregateID.Valid() || len(references.registered) != 1 ||
		len(references.released) != 0 || len(store.links) != 1 {
		t.Fatalf("create reference flow drifted: receipt=%+v refs=%+v links=%+v err=%v",
			receipt, references, store.links, err)
	}
	command, link := references.registered[0], store.links[0]
	if command.Link.TargetKind != "FORM" || command.Link.WorkResourceID != receipt.AggregateID ||
		link.ID != command.Link.ID || link.ReferenceEpoch != 7 ||
		link.TargetID != resource.Targets[0].ID || link.PermissionRevision != 13 {
		t.Fatalf("persisted link lacks exact owner receipt: command=%+v link=%+v",
			command, link)
	}
}

func TestCreateFormTaskStopsBeforeLinkPersistenceWhenRegistryFails(t *testing.T) {
	service, store, references := createResourceService(t, ports.ErrUnavailable)
	_, err := service.CreateTask(context.Background(), createScope{
		actor: updateID(8), tenant: updateID(9)}, ports.Mutation{
		IdempotencyKey: "create-form-task-failure"}, CreateTaskInput{
		BoardID: updateID(2), ColumnID: updateID(3), Kind: domain.TaskFormReview,
		Priority: domain.PriorityMedium, Title: "Review request",
		Resource: updateResourceInput(10),
	})
	if !errors.Is(err, ports.ErrUnavailable) || len(references.registered) != 1 ||
		len(store.links) != 0 {
		t.Fatalf("registry failure did not stop link persistence: refs=%+v links=%+v err=%v",
			references, store.links, err)
	}
}

func createResourceService(
	t *testing.T, referenceErr error,
) (*Service, *createTaskStore, *updateReferences) {
	t.Helper()
	boardID, columnID := updateID(2), updateID(3)
	tasks := &createTaskStore{}
	stores := createStores{boards: &createBoardStore{
		board: domain.Board{ID: boardID, Lifecycle: domain.LifecycleActive},
		column: domain.Column{ID: columnID, BoardID: boardID,
			Status: domain.StatusOpen}}, tasks: tasks}
	mutator := &updateResourceMutator{stores: stores, run: true}
	references := &updateReferences{err: referenceErr}
	service, err := New(Dependencies{Mutator: mutator,
		Principals: createPrincipalResolver{},
		Resources:  &updateResourceAuthority{allowed: true, revision: 13},
		References: references, Clock: updateClock{},
		IDs: &updateIDs{next: updateID(90)}})
	if err != nil {
		t.Fatalf("new create service: %v", err)
	}
	return service, tasks, references
}

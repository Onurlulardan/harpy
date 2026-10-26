package work

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/application/tasks"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

type legacyReplayTaskStore struct {
	ports.TaskStore
	task           domain.Task
	persists       int
	metadataWrites int
}

func (store *legacyReplayTaskStore) LoadTaskForUpdate(
	context.Context, domain.ID,
) (domain.Task, error) {
	return store.task, nil
}

func (store *legacyReplayTaskStore) PersistTask(
	_ context.Context, task domain.Task, expected int64,
) error {
	if store.task.RowVersion != expected {
		return ports.ErrConflict
	}
	store.persists++
	store.task = task
	return nil
}

func (store *legacyReplayTaskStore) UpsertLegacyTaskMetadata(
	context.Context, domain.ID, domain.LegacyTaskMetadata,
) error {
	store.metadataWrites++
	return nil
}

type legacyReplayOperations struct {
	ports.OperationStore
	hash         [32]byte
	receipt      ports.OperationReceipt
	acquired     bool
	completed    bool
	acquireCalls int
}

func (store *legacyReplayOperations) Acquire(_ context.Context, _ domain.ID,
	_, _ string, hash [32]byte,
) (ports.OperationReceipt, bool, error) {
	store.acquireCalls++
	if !store.acquired {
		store.acquired, store.hash = true, hash
		return ports.OperationReceipt{}, false, nil
	}
	if !bytes.Equal(store.hash[:], hash[:]) || !store.completed {
		return ports.OperationReceipt{}, false, ports.ErrConflict
	}
	return store.receipt, true, nil
}

func (*legacyReplayOperations) AppendAudit(
	context.Context, domain.ID, domain.ID, string, domain.ID, domain.ID, int64,
) (int64, error) {
	return 1, nil
}

func (store *legacyReplayOperations) PersistResult(_ context.Context, _ domain.ID,
	receipt ports.OperationReceipt, _ []byte,
) error {
	store.receipt, store.completed = receipt, true
	return nil
}

type legacyReplayStores struct {
	ports.Stores
	tasks      *legacyReplayTaskStore
	operations *legacyReplayOperations
}

func (stores legacyReplayStores) Tasks() ports.TaskStore           { return stores.tasks }
func (stores legacyReplayStores) Operations() ports.OperationStore { return stores.operations }

type legacyReplayUnit struct{ stores ports.Stores }

func (unit legacyReplayUnit) Bind(context.Context, ports.Scope) (ports.Stores, error) {
	return unit.stores, nil
}

type legacyReplayAuthorizer struct{}

func (legacyReplayAuthorizer) Authorize(context.Context, ports.Scope, string, domain.ID) error {
	return nil
}

type legacyReplayEvents struct{ records int }

func (events *legacyReplayEvents) Record(context.Context, ports.Scope, ports.EventIntent) error {
	events.records++
	return nil
}

type legacyReplayClock struct{}

func (legacyReplayClock) Now() time.Time {
	return time.Date(2026, 8, 2, 12, 0, 0, 0, time.UTC)
}

type legacyReplayIDs struct{ next byte }

func (ids *legacyReplayIDs) New(time.Time) (domain.ID, error) {
	ids.next++
	return workID(100 + ids.next), nil
}

type legacyReplayPrincipals struct{ ports.PrincipalResolver }
type legacyReplayResources struct{ ports.ResourceAuthority }
type legacyReplayReferences struct {
	ports.ResourceReferenceRegistry
}

func TestLegacyImplicitVersionUpdateReplaysExactReceipt(t *testing.T) {
	task := domain.Task{ID: workID(1), BoardID: workID(2), ColumnID: workID(3),
		CreatorID: workID(4), Kind: domain.TaskCustom, Status: domain.StatusOpen,
		Priority: domain.PriorityMedium, Title: "Recurring", Position: 1024, RowVersion: 1}
	taskStore := &legacyReplayTaskStore{task: task}
	operations := &legacyReplayOperations{}
	stores := legacyReplayStores{tasks: taskStore, operations: operations}
	events, ids := &legacyReplayEvents{}, &legacyReplayIDs{}
	mutator, err := NewMutationExecutor(MutationDependencies{UnitOfWork: legacyReplayUnit{stores},
		Authorizer: legacyReplayAuthorizer{}, Events: events, Clock: legacyReplayClock{}, IDs: ids})
	if err != nil {
		t.Fatal(err)
	}
	service, err := tasks.New(tasks.Dependencies{Mutator: mutator,
		Principals: legacyReplayPrincipals{}, Resources: legacyReplayResources{},
		References: legacyReplayReferences{}, Clock: legacyReplayClock{}, IDs: ids})
	if err != nil {
		t.Fatal(err)
	}
	body := []byte(`{"recurrence_rule":"FREQ=WEEKLY","form_filter":{"ignored":true}}`)
	intentHash, err := legacyTaskUpdateIntentHash(task.ID, body)
	if err != nil {
		t.Fatal(err)
	}
	input := tasks.UpdateInput{TaskID: task.ID, Priority: task.Priority, Title: task.Title,
		ReplaceLegacyMetadata: true, CompatibilityImplicitVersion: true,
		CompatibilityIntentHash: intentHash,
		LegacyMetadata: domain.LegacyTaskMetadata{IsRecurring: true,
			RecurrenceRule: "FREQ=WEEKLY"}}
	scope := taskResourceScope{actor: workID(8), tenant: workID(9)}
	first, err := service.UpdateTask(context.Background(), scope, ports.Mutation{
		IdempotencyKey: "legacy-replay", ExpectedVersion: 1}, input)
	if err != nil || first.RowVersion != 2 {
		t.Fatalf("first legacy update receipt=%+v err=%v", first, err)
	}
	// An unrelated successful task mutation changes both the row version and the
	// merged snapshot that the compatibility HTTP adapter would pass downstream.
	taskStore.task.Title = "Changed elsewhere"
	taskStore.task.RowVersion = 3
	replay := input
	replay.Title = taskStore.task.Title
	second, err := service.UpdateTask(context.Background(), scope, ports.Mutation{
		IdempotencyKey: "legacy-replay", ExpectedVersion: 3}, replay)
	if err != nil || second != first || taskStore.persists != 1 ||
		taskStore.metadataWrites != 1 || operations.acquireCalls != 2 || events.records != 1 {
		t.Fatalf("legacy replay receipt=%+v first=%+v persists=%d metadata=%d acquires=%d events=%d err=%v",
			second, first, taskStore.persists, taskStore.metadataWrites,
			operations.acquireCalls, events.records, err)
	}
	tamperedHash, err := legacyTaskUpdateIntentHash(task.ID,
		[]byte(`{"recurrence_rule":"FREQ=DAILY"}`))
	if err != nil {
		t.Fatal(err)
	}
	tampered := replay
	tampered.CompatibilityIntentHash = tamperedHash
	tampered.LegacyMetadata.RecurrenceRule = "FREQ=DAILY"
	if _, err = service.UpdateTask(context.Background(), scope, ports.Mutation{
		IdempotencyKey: "legacy-replay", ExpectedVersion: 3}, tampered); err != ports.ErrConflict {
		t.Fatalf("same key with changed legacy intent error = %v", err)
	}
	if taskStore.persists != 1 || taskStore.metadataWrites != 1 || events.records != 1 {
		t.Fatalf("tampered replay mutated state: persists=%d metadata=%d events=%d",
			taskStore.persists, taskStore.metadataWrites, events.records)
	}
}

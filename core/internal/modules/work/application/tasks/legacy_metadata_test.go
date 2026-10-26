package tasks

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func TestCreateTaskPersistsLegacyMetadataInMutation(t *testing.T) {
	service, store, _ := createResourceService(t, nil)
	end := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	_, err := service.CreateTask(context.Background(), createScope{
		actor: updateID(8), tenant: updateID(9)}, ports.Mutation{
		IdempotencyKey: "create-legacy-metadata"}, CreateTaskInput{
		BoardID: updateID(2), ColumnID: updateID(3), Kind: domain.TaskCustom,
		Priority: domain.PriorityMedium, Title: "Recurring metadata",
		LegacyMetadata: &domain.LegacyTaskMetadata{IsRecurring: true,
			RecurrenceRule: "FREQ=WEEKLY", RecurrenceEnd: &end,
			FormFilter: []byte(`{"status":"open"}`)},
	})
	if err != nil || store.legacyWrites != 1 || !store.legacy.IsRecurring ||
		store.legacy.RecurrenceRule != "FREQ=WEEKLY" ||
		string(store.legacy.FormFilter) != `{"status":"open"}` {
		t.Fatalf("legacy create metadata drifted: value=%+v writes=%d err=%v",
			store.legacy, store.legacyWrites, err)
	}
}

func TestUpdateTaskPersistsLegacyMetadataAndBindsFingerprint(t *testing.T) {
	task := validUpdateTask(domain.TaskCustom)
	store := &updateResourceStore{task: task}
	mutator := &updateResourceMutator{stores: updateResourceStores{tasks: store}, run: true}
	service := updateService(mutator, &updateResourceAuthority{allowed: true})
	input := updateInput(task)
	input.ReplaceLegacyMetadata = true
	input.LegacyMetadata = domain.LegacyTaskMetadata{IsRecurring: true,
		RecurrenceRule: "FREQ=DAILY", FormFilter: []byte(`{}`)}
	_, err := service.UpdateTask(context.Background(), nil, ports.Mutation{
		ExpectedVersion: task.RowVersion, IdempotencyKey: "legacy-update"}, input)
	if err != nil || store.legacyWrites != 1 || !store.legacy.IsRecurring {
		t.Fatalf("legacy update metadata drifted: value=%+v writes=%d err=%v",
			store.legacy, store.legacyWrites, err)
	}

	mutator.run = false
	input.LegacyMetadata.IsRecurring = false
	_, _ = service.UpdateTask(context.Background(), nil, ports.Mutation{
		ExpectedVersion: task.RowVersion, IdempotencyKey: "legacy-update-2"}, input)
	if len(mutator.hashes) != 2 || mutator.hashes[0] == mutator.hashes[1] {
		t.Fatal("legacy metadata was omitted from the idempotency fingerprint")
	}
}

func TestUpdateTaskReturnsMetadataPersistenceFailure(t *testing.T) {
	task := validUpdateTask(domain.TaskCustom)
	store := &updateResourceStore{task: task, legacyErr: ports.ErrUnavailable}
	mutator := &updateResourceMutator{stores: updateResourceStores{tasks: store}, run: true}
	input := updateInput(task)
	input.ReplaceLegacyMetadata = true
	input.LegacyMetadata = domain.LegacyTaskMetadata{FormFilter: []byte(`{}`)}
	_, err := updateService(mutator, &updateResourceAuthority{allowed: true}).UpdateTask(
		context.Background(), nil, ports.Mutation{ExpectedVersion: task.RowVersion,
			IdempotencyKey: "legacy-failure"}, input)
	if !errors.Is(err, ports.ErrUnavailable) || store.legacyWrites != 1 {
		t.Fatalf("metadata storage failure was hidden: writes=%d err=%v",
			store.legacyWrites, err)
	}
}

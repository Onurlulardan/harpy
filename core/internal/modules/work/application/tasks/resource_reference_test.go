package tasks

import (
	"context"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func TestUpdateTaskRegistersOnlyDefinitionReference(t *testing.T) {
	store := &updateResourceStore{task: validUpdateTask(domain.TaskCustom)}
	mutator := &updateResourceMutator{stores: updateResourceStores{tasks: store}, run: true}
	references := &updateReferences{}
	resource := updateResourceInput(26)
	resource.Targets = append(resource.Targets, ResourceTargetInput{Kind: "FORM_RECORD",
		ID: updateID(29), VersionID: resource.Targets[0].VersionID})
	input := updateInput(store.task)
	input.ReplaceResource, input.Kind, input.Resource = true, domain.TaskFormFill, resource
	_, err := updateServiceWithReferences(mutator,
		&updateResourceAuthority{allowed: true, revision: 13}, references).UpdateTask(
		context.Background(), nil, ports.Mutation{ExpectedVersion: 3}, input)
	if err != nil || len(references.registered) != 1 || len(references.released) != 0 ||
		references.registered[0].Link.TargetKind != "FORM" ||
		references.registered[0].OperationID == references.registered[0].Link.ID ||
		references.registered[0].OccurredAt.Nanosecond()%1_000 != 0 {
		t.Fatalf("owner reference registration drifted: %+v err=%v", references, err)
	}
}

func TestUpdateTaskReleasesExactPreviousDefinitionReference(t *testing.T) {
	store := &updateResourceStore{task: validUpdateTask(domain.TaskFormReview)}
	store.currentLinks = storedUpdateResource(store.task, 82)
	mutator := &updateResourceMutator{stores: updateResourceStores{tasks: store}, run: true}
	references := &updateReferences{}
	input := updateInput(store.task)
	input.ReplaceResource, input.Kind = true, domain.TaskCustom
	_, err := updateServiceWithReferences(mutator,
		&updateResourceAuthority{allowed: true, revision: 1}, references).UpdateTask(
		context.Background(), nil, ports.Mutation{ExpectedVersion: 3}, input)
	if err != nil || len(references.released) != 1 || len(references.registered) != 0 ||
		references.released[0].Link != store.currentLinks[0] {
		t.Fatalf("owner reference release drifted: %+v err=%v", references, err)
	}
}

func TestUpdateTaskFailsWhenOwnerReferenceRegistrationFails(t *testing.T) {
	store := &updateResourceStore{task: validUpdateTask(domain.TaskCustom)}
	mutator := &updateResourceMutator{stores: updateResourceStores{tasks: store}, run: true}
	references := &updateReferences{err: ports.ErrUnavailable}
	input := updateInput(store.task)
	input.ReplaceResource, input.Kind, input.Resource = true, domain.TaskFormFill,
		updateResourceInput(34)
	_, err := updateServiceWithReferences(mutator,
		&updateResourceAuthority{allowed: true, revision: 2}, references).UpdateTask(
		context.Background(), nil, ports.Mutation{ExpectedVersion: 3}, input)
	if err == nil || store.persisted.Valid() || len(references.registered) != 1 {
		t.Fatalf("failed owner registration was accepted: persisted=%+v refs=%+v err=%v",
			store.persisted, references, err)
	}
}

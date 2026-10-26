package tasks

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func TestUpdateTaskPreservesResourceWithoutExplicitReplacement(t *testing.T) {
	store := &updateResourceStore{task: validUpdateTask(domain.TaskFormFill)}
	mutator := &updateResourceMutator{stores: updateResourceStores{tasks: store}, run: true}
	authority := &updateResourceAuthority{allowed: true, revision: 9}
	input := updateInput(store.task)
	_, err := updateService(mutator, authority).UpdateTask(context.Background(), nil,
		ports.Mutation{ExpectedVersion: 3}, input)
	if err != nil || store.replaced || store.persisted.Kind != domain.TaskFormFill ||
		authority.calls != 0 {
		t.Fatalf("resource was not preserved: replaced=%v kind=%s calls=%d err=%v",
			store.replaced, store.persisted.Kind, authority.calls, err)
	}
}

func TestUpdateTaskAtomicallyReplacesFormResource(t *testing.T) {
	store := &updateResourceStore{task: validUpdateTask(domain.TaskCustom)}
	mutator := &updateResourceMutator{stores: updateResourceStores{tasks: store}, run: true}
	authority := &updateResourceAuthority{allowed: true, revision: 11}
	input := updateInput(store.task)
	input.ReplaceResource, input.Kind, input.Resource = true, domain.TaskFormReview,
		updateResourceInput(20)
	_, err := updateService(mutator, authority).UpdateTask(context.Background(), nil,
		ports.Mutation{ExpectedVersion: 3}, input)
	var link *domain.ResourceLink
	if len(store.replacements) == 1 {
		link = &store.replacements[0]
	}
	if err != nil || !store.replaced || store.persisted.Kind != domain.TaskFormReview ||
		link == nil || link.WorkResourceID != store.task.ID ||
		link.TargetID != input.Resource.Targets[0].ID || link.PermissionRevision != 11 ||
		authority.calls != 1 {
		t.Fatalf("resource replacement drifted: task=%+v link=%+v calls=%d err=%v",
			store.persisted, link, authority.calls, err)
	}
}

func TestUpdateTaskPersistsDefinitionAndExactRecordTargets(t *testing.T) {
	store := &updateResourceStore{task: validUpdateTask(domain.TaskCustom)}
	mutator := &updateResourceMutator{stores: updateResourceStores{tasks: store}, run: true}
	authority := &updateResourceAuthority{allowed: true, revision: 12}
	resource := updateResourceInput(21)
	resource.Targets = append(resource.Targets, ResourceTargetInput{Kind: "FORM_RECORD",
		ID: updateID(24), VersionID: resource.Targets[0].VersionID})
	input := updateInput(store.task)
	input.ReplaceResource, input.Kind, input.Resource = true, domain.TaskFormFill, resource
	_, err := updateService(mutator, authority).UpdateTask(context.Background(), nil,
		ports.Mutation{ExpectedVersion: 3}, input)
	if err != nil || len(store.replacements) != 2 || authority.calls != 2 ||
		store.replacements[0].TargetKind != "FORM" ||
		store.replacements[1].TargetKind != "FORM_RECORD" ||
		store.replacements[1].TargetID != updateID(24) ||
		store.replacements[0].ID == store.replacements[1].ID {
		t.Fatalf("definition/record resource set drifted: %+v calls=%d err=%v",
			store.replacements, authority.calls, err)
	}
}

func TestUpdateTaskSwitchToCustomReleasesFormResource(t *testing.T) {
	store := &updateResourceStore{task: validUpdateTask(domain.TaskFormFill)}
	store.currentLinks = storedUpdateResource(store.task, 80)
	mutator := &updateResourceMutator{stores: updateResourceStores{tasks: store}, run: true}
	authority := &updateResourceAuthority{allowed: true, revision: 11}
	input := updateInput(store.task)
	input.ReplaceResource, input.Kind = true, domain.TaskCustom
	_, err := updateService(mutator, authority).UpdateTask(context.Background(), nil,
		ports.Mutation{ExpectedVersion: 3}, input)
	if err != nil || !store.replaced || len(store.replacements) != 0 ||
		store.persisted.Kind != domain.TaskCustom || authority.calls != 0 {
		t.Fatalf("resource release drifted: task=%+v replacement=%+v calls=%d err=%v",
			store.persisted, store.replacements, authority.calls, err)
	}
}

func TestUpdateTaskFailsClosedBeforeMutationWhenResourceIsDenied(t *testing.T) {
	mutator := &updateResourceMutator{}
	authority := &updateResourceAuthority{allowed: false, revision: 0,
		err: ports.ErrUnavailable}
	task := validUpdateTask(domain.TaskCustom)
	input := updateInput(task)
	input.ReplaceResource, input.Kind, input.Resource = true, domain.TaskFormFill,
		updateResourceInput(30)
	_, err := updateService(mutator, authority).UpdateTask(context.Background(), nil,
		ports.Mutation{ExpectedVersion: 3}, input)
	if !errors.Is(err, ports.ErrForbidden) || len(mutator.hashes) != 0 {
		t.Fatalf("denied resource reached mutation: hashes=%d err=%v", len(mutator.hashes), err)
	}
}

func TestUpdateTaskFingerprintBindsResourceReplacement(t *testing.T) {
	mutator := &updateResourceMutator{}
	authority := &updateResourceAuthority{allowed: true, revision: 5}
	service, task := updateService(mutator, authority), validUpdateTask(domain.TaskCustom)
	for _, resource := range []*ResourceLinkInput{updateResourceInput(40), updateResourceInput(50)} {
		input := updateInput(task)
		input.ReplaceResource, input.Kind, input.Resource = true, domain.TaskFormFill, resource
		if _, err := service.UpdateTask(context.Background(), nil,
			ports.Mutation{ExpectedVersion: 3}, input); err != nil {
			t.Fatal(err)
		}
	}
	if len(mutator.hashes) != 2 || mutator.hashes[0] == mutator.hashes[1] {
		t.Fatalf("resource identity was absent from idempotency fingerprint: %x %x",
			mutator.hashes[0], mutator.hashes[1])
	}
}

func TestUpdateTaskFingerprintBindsCurrentPermissionRevision(t *testing.T) {
	mutator := &updateResourceMutator{}
	authority := &updateResourceAuthority{allowed: true, revision: 5}
	service, task := updateService(mutator, authority), validUpdateTask(domain.TaskCustom)
	resource := updateResourceInput(55)
	for _, revision := range []int64{5, 6} {
		authority.revision = revision
		input := updateInput(task)
		input.ReplaceResource, input.Kind, input.Resource = true,
			domain.TaskFormFill, resource
		if _, err := service.UpdateTask(context.Background(), nil,
			ports.Mutation{ExpectedVersion: 3}, input); err != nil {
			t.Fatal(err)
		}
	}
	if len(mutator.hashes) != 2 || mutator.hashes[0] == mutator.hashes[1] {
		t.Fatalf("permission revision was absent from fingerprint: %x %x",
			mutator.hashes[0], mutator.hashes[1])
	}
}

func TestUpdateTaskRejectsAmbiguousResourceInput(t *testing.T) {
	mutator := &updateResourceMutator{}
	task := validUpdateTask(domain.TaskCustom)
	input := updateInput(task)
	input.Resource = updateResourceInput(60)
	_, err := updateService(mutator, &updateResourceAuthority{}).UpdateTask(
		context.Background(), nil, ports.Mutation{ExpectedVersion: 3}, input)
	if !errors.Is(err, ports.ErrInvalid) {
		t.Fatalf("ambiguous resource input accepted: %v", err)
	}
}

func TestUpdateTaskRejectsNonCanonicalResourceTargets(t *testing.T) {
	mutator := &updateResourceMutator{}
	task := validUpdateTask(domain.TaskCustom)
	resource := updateResourceInput(70)
	resource.Targets = append([]ResourceTargetInput{{Kind: "FORM_RECORD",
		ID: updateID(73), VersionID: resource.Targets[0].VersionID}},
		resource.Targets...)
	input := updateInput(task)
	input.ReplaceResource, input.Kind, input.Resource = true, domain.TaskFormFill, resource
	_, err := updateService(mutator, &updateResourceAuthority{allowed: true, revision: 1}).UpdateTask(
		context.Background(), nil, ports.Mutation{ExpectedVersion: 3}, input)
	if !errors.Is(err, ports.ErrInvalid) || len(mutator.hashes) != 0 {
		t.Fatalf("non-canonical resource targets accepted: %v", err)
	}
}

package work

import (
	"context"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/application/tasks"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

type taskResourceScope struct{ actor, tenant domain.ID }

func (scope taskResourceScope) TenantID() domain.ID { return scope.tenant }
func (scope taskResourceScope) ActorID() domain.ID  { return scope.actor }
func (taskResourceScope) OperationID() string       { return "work.task.update" }
func (taskResourceScope) SnapshotQuery() bool       { return false }

type taskResourceResolverFixture struct {
	resolved      *tasks.ResourceLinkInput
	description   LegacyTaskResourceDescription
	resolveInput  LegacyTaskResourceRequest
	resolveCalls  int
	describeCalls int
}

func (fixture *taskResourceResolverFixture) Resolve(
	_ context.Context, _ ports.Scope, request LegacyTaskResourceRequest,
) (*tasks.ResourceLinkInput, error) {
	fixture.resolveCalls++
	fixture.resolveInput = request
	return fixture.resolved, nil
}

func (fixture *taskResourceResolverFixture) Describe(
	_ context.Context, _ ports.Scope, _ *tasks.ResourceLinkInput,
) (LegacyTaskResourceDescription, error) {
	fixture.describeCalls++
	return fixture.description, nil
}

type taskResourceLinksFixture struct {
	links []domain.ResourceLink
	calls int
}

func (fixture *taskResourceLinksFixture) LoadResourceLinksForUpdate(
	_ context.Context, _ domain.ID,
) ([]domain.ResourceLink, error) {
	fixture.calls++
	return append([]domain.ResourceLink(nil), fixture.links...), nil
}

func TestCreateFormTaskResolvesLegacyAliasThroughOwner(t *testing.T) {
	resource := formResourceInput(workID(3), workID(4), nil)
	resolver := &taskResourceResolverFixture{resolved: resource}
	runtime := &PlatformHTTPRuntime{taskResources: resolver}
	actual, err := runtime.createTaskResource(context.Background(), taskScope(), true,
		ports.TaskFormFill, writeWire{FormSlug: "leave-request",
			FormDataID: workID(5).String()})
	if err != nil || actual != resource || resolver.resolveCalls != 1 ||
		resolver.resolveInput.FormSlug != "leave-request" ||
		resolver.resolveInput.FormDataID != workID(5).String() {
		t.Fatalf("owner resolution drifted: actual=%+v request=%+v err=%v",
			actual, resolver.resolveInput, err)
	}
}

func TestCreateTaskRejectsUnownedOrInconsistentFormIdentity(t *testing.T) {
	runtime := &PlatformHTTPRuntime{taskResources: &taskResourceResolverFixture{}}
	if _, err := runtime.createTaskResource(context.Background(), taskScope(), false,
		ports.TaskFormFill, writeWire{FormSlug: "leave-request"}); err == nil {
		t.Fatal("canonical form task without canonical resource contract was accepted")
	}
	if _, err := runtime.createTaskResource(context.Background(), taskScope(), true,
		ports.TaskCustom, writeWire{FormSlug: "leave-request"}); err == nil {
		t.Fatal("custom task with a form identity was accepted")
	}
}

func TestUpdateFormKindPreservesExactStoredTargets(t *testing.T) {
	taskID, definitionID, versionID := workID(1), workID(3), workID(4)
	links := &taskResourceLinksFixture{links: formResourceLinks(
		taskID, definitionID, versionID, nil)}
	resolver := &taskResourceResolverFixture{}
	runtime := &PlatformHTTPRuntime{taskResources: resolver}
	replace, kind, resource, err := runtime.updateTaskResource(context.Background(),
		taskScope(), links, domain.Task{ID: taskID, Kind: ports.TaskFormFill}, true,
		writeWire{Present: map[string]bool{"task_type": true},
			TaskType: string(ports.TaskFormReview)})
	if err != nil || !replace || kind != ports.TaskFormReview || resource == nil ||
		resource.Targets[0].ID != definitionID || resolver.resolveCalls != 0 ||
		resolver.describeCalls != 0 || links.calls != 1 {
		t.Fatalf("exact target was not preserved: replace=%v kind=%s resource=%+v err=%v",
			replace, kind, resource, err)
	}
}

func TestUpdateFormRecordClearUsesDescriptionThenOwnerResolution(t *testing.T) {
	taskID, definitionID, versionID, recordID := workID(1), workID(3), workID(4), workID(5)
	links := &taskResourceLinksFixture{links: formResourceLinks(
		taskID, definitionID, versionID, &recordID)}
	replacement := formResourceInput(definitionID, versionID, nil)
	resolver := &taskResourceResolverFixture{resolved: replacement,
		description: LegacyTaskResourceDescription{FormSlug: "leave-request",
			FormDataID: recordID.String()}}
	runtime := &PlatformHTTPRuntime{taskResources: resolver}
	replace, kind, resource, err := runtime.updateTaskResource(context.Background(),
		taskScope(), links, domain.Task{ID: taskID, Kind: ports.TaskFormFill}, true,
		writeWire{Present: map[string]bool{"form_data_id": true}})
	if err != nil || !replace || kind != ports.TaskFormFill || resource != replacement ||
		resolver.describeCalls != 1 || resolver.resolveCalls != 1 ||
		resolver.resolveInput.FormSlug != "leave-request" ||
		resolver.resolveInput.FormDataID != "" {
		t.Fatalf("record clear drifted: request=%+v err=%v",
			resolver.resolveInput, err)
	}
}

func TestUpdateFormTaskToCustomReleasesCanonicalResource(t *testing.T) {
	runtime := &PlatformHTTPRuntime{taskResources: &taskResourceResolverFixture{}}
	replace, kind, resource, err := runtime.updateTaskResource(context.Background(),
		taskScope(), &taskResourceLinksFixture{},
		domain.Task{ID: workID(1), Kind: ports.TaskFormReview}, true,
		writeWire{Present: map[string]bool{"task_type": true},
			TaskType: string(ports.TaskCustom)})
	if err != nil || !replace || kind != ports.TaskCustom || resource != nil {
		t.Fatalf("form-to-custom transition drifted: %v %s %+v %v",
			replace, kind, resource, err)
	}
}

func TestUpdateFormTaskIgnoresLegacyNullPointerFields(t *testing.T) {
	runtime := &PlatformHTTPRuntime{taskResources: &taskResourceResolverFixture{}}
	replace, kind, resource, err := runtime.updateTaskResource(context.Background(),
		taskScope(), &taskResourceLinksFixture{},
		domain.Task{ID: workID(1), Kind: ports.TaskFormReview}, true,
		writeWire{Present: map[string]bool{"task_type": true, "form_slug": true,
			"form_data_id": true}, Null: map[string]bool{"task_type": true,
			"form_slug": true, "form_data_id": true}})
	if err != nil || replace || kind != "" || resource != nil {
		t.Fatalf("legacy null form patch mutated resource: %v %q %+v %v",
			replace, kind, resource, err)
	}
}

func taskScope() ports.Scope {
	return taskResourceScope{actor: workID(8), tenant: workID(9)}
}

func workID(seed byte) domain.ID {
	var value domain.ID
	value[0], value[6], value[8] = seed, 0x70, 0x80
	return value
}

func formResourceInput(
	definitionID, versionID domain.ID, recordID *domain.ID,
) *tasks.ResourceLinkInput {
	result := &tasks.ResourceLinkInput{Owner: domain.SourceForms,
		Targets: []tasks.ResourceTargetInput{{Kind: "FORM", ID: definitionID,
			VersionID: versionID}}}
	if recordID != nil {
		result.Targets = append(result.Targets, tasks.ResourceTargetInput{
			Kind: "FORM_RECORD", ID: *recordID, VersionID: versionID})
	}
	return result
}

func formResourceLinks(
	taskID, definitionID, versionID domain.ID, recordID *domain.ID,
) []domain.ResourceLink {
	links := []domain.ResourceLink{{ID: workID(6), WorkResourceID: taskID,
		TargetOwner: domain.SourceForms, TargetKind: "FORM", TargetID: definitionID,
		TargetVersionID: versionID, ReferenceEpoch: 2,
		PermissionRevision: 3, RowVersion: 1}}
	if recordID != nil {
		links = append(links, domain.ResourceLink{ID: workID(7), WorkResourceID: taskID,
			TargetOwner: domain.SourceForms, TargetKind: "FORM_RECORD", TargetID: *recordID,
			TargetVersionID: versionID, ReferenceEpoch: 2,
			PermissionRevision: 3, RowVersion: 1})
	}
	return links
}

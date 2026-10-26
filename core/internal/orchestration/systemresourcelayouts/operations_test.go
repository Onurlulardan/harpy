package systemresourcelayouts

import (
	"context"
	"slices"
	"testing"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestLayoutRevisionOperationsStayInOneOwnedScope(t *testing.T) {
	fixture := newFixture(t, publicresource.Users)
	ctx := context.Background()
	definition := emptyLayoutDefinition()
	if _, err := fixture.application.Create(
		ctx, fixture.admission, publicresource.Users,
		fixture.layout.revision.ID, definition,
	); err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.application.Get(
		ctx, fixture.admission, publicresource.Users, fixture.layout.revision.ID,
	); err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.application.List(
		ctx, fixture.admission, publicresource.Users,
	); err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.application.Update(
		ctx, fixture.admission, publicresource.Users,
		layoutapi.UpdateInput{
			ID: fixture.layout.revision.ID, ExpectedRowRevision: 1,
			Definition: definition,
		},
	); err != nil {
		t.Fatal(err)
	}
	if err := fixture.application.Delete(
		ctx, fixture.admission, publicresource.Users,
		layoutapi.DeleteInput{ID: fixture.layout.revision.ID, ExpectedRowRevision: 1},
	); err != nil {
		t.Fatal(err)
	}
	if fixture.coordinator.calls != 5 || fixture.layout.createCalls != 1 ||
		fixture.layout.createInput.ID != fixture.layout.revision.ID ||
		fixture.layout.getCalls != 3 || fixture.layout.listCalls != 1 ||
		fixture.layout.updateCalls != 1 || fixture.layout.deleteCalls != 1 ||
		!slices.Equal(fixture.coordinator.modes, []platformtx.ScopeMode{
			platformtx.Command, platformtx.SnapshotQuery, platformtx.SnapshotQuery,
			platformtx.Command, platformtx.Command,
		}) {
		t.Fatalf("unexpected calls: coordinator=%d layout=%+v", fixture.coordinator.calls, fixture.layout)
	}
}

func TestResolveUsesCurrentIAMCatalogAndLayoutCompiler(t *testing.T) {
	for _, resource := range []publicresource.Resource{
		publicresource.Users, publicresource.Roles, publicresource.Organizations,
	} {
		t.Run(string(resource), func(t *testing.T) {
			fixture := newFixture(t, resource)
			document, err := fixture.application.Resolve(
				context.Background(), fixture.admission, resource,
			)
			if err != nil || document.RegistryContract == "" ||
				fixture.coordinator.calls != 1 || fixture.iam.calls != 1 ||
				fixture.layout.resolveCalls != 1 ||
				fixture.coordinator.modes[0] != platformtx.SnapshotQuery {
				t.Fatalf("resolve=%+v iam=%d layout=%+v: %v",
					document, fixture.iam.calls, fixture.layout, err)
			}
		})
	}
}

func emptyLayoutDefinition() layoutapi.Definition {
	kinds := []layoutapi.SurfaceKind{
		layoutapi.SurfaceRecordCreate, layoutapi.SurfaceRecordUpdate,
		layoutapi.SurfaceRecordDetail, layoutapi.SurfaceRecordList,
		layoutapi.SurfaceRecordFilter,
	}
	surfaces := make([]layoutapi.SurfaceDefinition, len(kinds))
	for index, kind := range kinds {
		surfaces[index] = layoutapi.SurfaceDefinition{
			Kind: kind, Sections: []layoutapi.Section{},
			Translations: []layoutapi.Translation{},
		}
	}
	return layoutapi.Definition{
		SourceContract: layoutapi.DefinitionContractV2, Surfaces: surfaces,
	}
}

func TestLayoutRevisionRejectsCrossResourceIdentity(t *testing.T) {
	fixture := newFixture(t, publicresource.Users)
	_, err := fixture.application.Get(
		context.Background(), fixture.admission,
		publicresource.Roles, fixture.layout.revision.ID,
	)
	if err != layoutapi.ErrNotFound || fixture.layout.getCalls != 1 {
		t.Fatalf("cross-resource result = %v calls=%d", err, fixture.layout.getCalls)
	}
}

func TestPublishUsesCurrentIAMCatalogInSameScope(t *testing.T) {
	for _, resource := range []publicresource.Resource{
		publicresource.Users, publicresource.Roles, publicresource.Organizations,
	} {
		t.Run(string(resource), func(t *testing.T) {
			fixture := newFixture(t, resource)
			result, err := fixture.application.Publish(
				context.Background(), fixture.admission,
				PublishInput{
					Resource:   resource,
					RevisionID: fixture.layout.revision.ID, ExpectedRowRevision: 1,
				},
			)
			if err != nil {
				t.Fatal(err)
			}
			if result.State != layoutapi.RevisionPublished ||
				fixture.coordinator.calls != 1 || fixture.iam.calls != 0 ||
				fixture.iam.publishCatalogCalls != 1 ||
				fixture.layout.getCalls != 1 || fixture.layout.publishCalls != 1 ||
				fixture.coordinator.modes[0] != platformtx.Command {
				t.Fatalf("publish result=%+v iam=%d publish_iam=%d layout=%+v",
					result, fixture.iam.calls,
					fixture.iam.publishCatalogCalls, fixture.layout)
			}
		})
	}
}

func TestConstructorRejectsMissingOwnerDependency(t *testing.T) {
	fixture := newFixture(t, publicresource.Users)
	if _, err := NewApplication(nil, fixture.iam, fixture.layout); err != ErrContract {
		t.Fatalf("nil coordinator error = %v", err)
	}
	if _, err := NewApplication(fixture.coordinator, nil, fixture.layout); err != ErrContract {
		t.Fatalf("nil IAM error = %v", err)
	}
	if _, err := NewApplication(
		fixture.coordinator, &testIAMWithoutLayoutCatalog{}, fixture.layout,
	); err != ErrContract {
		t.Fatalf("missing layout catalog contract error = %v", err)
	}
	if _, err := NewApplication(fixture.coordinator, fixture.iam, nil); err != ErrContract {
		t.Fatalf("nil layout error = %v", err)
	}
}

package systemresourcelayouts

import (
	"context"
	"crypto/sha256"
	"testing"
	"time"

	fieldregistry "github.com/Onurlulardan/harpy/core/internal/platform/fieldcatalog/fieldregistry"
	formpresentation "github.com/Onurlulardan/harpy/core/internal/platform/fieldcatalog/presentation"
	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type testCoordinator struct {
	calls      int
	activeMode platformtx.ScopeMode
	modes      []platformtx.ScopeMode
}

func (coordinator *testCoordinator) InTenant(
	ctx context.Context,
	request platformtx.ScopeRequest,
	execute func(context.Context, platformtx.TenantScope) error,
) error {
	coordinator.calls++
	coordinator.activeMode = request.Mode
	coordinator.modes = append(coordinator.modes, request.Mode)
	defer func() { coordinator.activeMode = "" }()
	return execute(ctx, platformtx.TenantScope{})
}

type testIAM struct {
	publicresource.Port
	catalog             publicresource.FieldCatalog
	coordinator         *testCoordinator
	calls               int
	publishCatalogCalls int
}

type testIAMWithoutLayoutCatalog struct{ publicresource.Port }

func (iam *testIAM) Catalog(
	context.Context,
	platformtx.TenantScope,
	publicresource.ActorEvidence,
	publicresource.Resource,
) (publicresource.FieldCatalog, error) {
	if iam.coordinator == nil ||
		iam.coordinator.activeMode != platformtx.SnapshotQuery {
		return publicresource.FieldCatalog{}, publicresource.ErrContract
	}
	iam.calls++
	return iam.catalog, nil
}

func (iam *testIAM) CatalogForLayoutPublish(
	context.Context,
	platformtx.TenantScope,
	publicresource.ActorEvidence,
	publicresource.Resource,
) (publicresource.FieldCatalog, error) {
	if iam.coordinator == nil ||
		iam.coordinator.activeMode != platformtx.Command {
		return publicresource.FieldCatalog{}, publicresource.ErrContract
	}
	iam.publishCatalogCalls++
	return iam.catalog, nil
}

type testLayout struct {
	revision     layoutapi.Revision
	document     layoutapi.PresentationDocument
	createInput  layoutapi.CreateInput
	createCalls  int
	getCalls     int
	listCalls    int
	updateCalls  int
	deleteCalls  int
	publishCalls int
	resolveCalls int
}

func (layout *testLayout) CreateLayout(
	_ context.Context, _ platformtx.TenantScope, input layoutapi.CreateInput,
) (layoutapi.Revision, error) {
	layout.createCalls++
	layout.createInput = input
	return layout.revision, nil
}

func (layout *testLayout) GetLayout(
	context.Context, platformtx.TenantScope, string,
) (layoutapi.Revision, error) {
	layout.getCalls++
	return layout.revision, nil
}

func (layout *testLayout) ListLayouts(
	context.Context, platformtx.TenantScope, layoutapi.ResourceKind,
) ([]layoutapi.Revision, error) {
	layout.listCalls++
	return []layoutapi.Revision{layout.revision}, nil
}

func (layout *testLayout) UpdateLayout(
	context.Context, platformtx.TenantScope, layoutapi.UpdateInput,
) (layoutapi.Revision, error) {
	layout.updateCalls++
	return layout.revision, nil
}

func (layout *testLayout) DeleteLayout(
	context.Context, platformtx.TenantScope, layoutapi.DeleteInput,
) error {
	layout.deleteCalls++
	return nil
}

func (layout *testLayout) PublishPresentation(
	context.Context, platformtx.TenantScope, layoutapi.PublishPresentationInput,
) (layoutapi.Revision, error) {
	layout.publishCalls++
	result := layout.revision
	result.State = layoutapi.RevisionPublished
	return result, nil
}

func (layout *testLayout) ResolvePresentation(
	context.Context, platformtx.TenantScope, layoutapi.ResourceKind,
	layoutapi.PresentationFieldCatalog,
) (layoutapi.PresentationDocument, error) {
	layout.resolveCalls++
	if layout.document.RegistryContract != "" {
		return layout.document, nil
	}
	return layoutapi.PresentationDocument{
		RegistryContract: formpresentation.RegistryContractVersion,
		DefaultLocale:    "en", Locales: []string{"en"},
		Surfaces: []formpresentation.Surface{},
	}, nil
}

type testFixture struct {
	application *Application
	coordinator *testCoordinator
	iam         *testIAM
	layout      *testLayout
	admission   publicresource.Admission
}

func newFixture(t *testing.T, resource publicresource.Resource) testFixture {
	t.Helper()
	tenantID, _ := publicresource.ParseID("018f47f4-7b1b-7cc0-8000-000000000001")
	actorID, _ := publicresource.ParseID("018f47f4-7b1b-7cc0-8000-000000000002")
	actor, _ := publicresource.NewActorEvidence(tenantID, actorID, 1, 1)
	admission, _ := publicresource.NewAdmission(tenantID, actor, 7, 7, 7)
	kind, _ := layoutResource(resource)
	revision := layoutapi.Revision{
		ID: "018f47f4-7b1b-7cc0-8000-000000000003", Resource: kind,
		Version: 1, State: layoutapi.RevisionDraft, RowRevision: 1,
		CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	}
	layout := &testLayout{revision: revision}
	coordinator := &testCoordinator{}
	iam := &testIAM{catalog: testCatalog(t, resource), coordinator: coordinator}
	application, err := NewApplication(coordinator, iam, layout)
	if err != nil {
		t.Fatal(err)
	}
	return testFixture{application, coordinator, iam, layout, admission}
}

func testCatalog(t *testing.T, resource publicresource.Resource) publicresource.FieldCatalog {
	t.Helper()
	field, _ := formpresentation.NewField(
		"a1000000-0000-4000-8000-000000000001",
		fieldregistry.TypeUUID, fieldregistry.ValueUUIDRFC4122LowerV1,
		formpresentation.FieldReadOnly,
	)
	presentation, _ := formpresentation.NewFieldCatalog([]formpresentation.Field{field})
	resourceID, _ := resource.CatalogResourceID()
	parsedResourceID, _ := publicresource.ParseID(resourceID)
	versionID, _ := publicresource.ParseID("018f47f4-7b1b-7cc0-8000-000000000004")
	fieldID, _ := publicresource.ParseID("a1000000-0000-4000-8000-000000000001")
	descriptor, _ := publicresource.NewFieldDescriptor(fieldID, "id")
	catalog, err := publicresource.NewFieldCatalog(
		resource, parsedResourceID, versionID, sha256.Sum256([]byte("catalog")), presentation,
		[]publicresource.FieldDescriptor{descriptor},
	)
	if err != nil {
		t.Fatal(err)
	}
	return catalog
}

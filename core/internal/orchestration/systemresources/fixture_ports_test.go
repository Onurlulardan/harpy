package systemresources

import (
	"context"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type testCoordinator struct {
	calls     int
	committed bool
}

func (coordinator *testCoordinator) InTenant(
	ctx context.Context,
	_ platformtx.ScopeRequest,
	execute func(context.Context, platformtx.TenantScope) error,
) error {
	coordinator.calls++
	err := execute(ctx, platformtx.TenantScope{})
	coordinator.committed = err == nil
	return err
}

type testIAMPort struct {
	fixture       testFixture
	preparedCalls int
	scope         platformtx.TenantScope
}

func (port *testIAMPort) Catalog(
	_ context.Context, scope platformtx.TenantScope,
	_ publicresource.ActorEvidence, _ publicresource.Resource,
) (publicresource.FieldCatalog, error) {
	port.scope = scope
	return port.fixture.catalog, nil
}

func (port *testIAMPort) PrepareCreate(
	_ context.Context, _ publicresource.CreateInput,
) (publicresource.PreparedCreate, error) {
	port.preparedCalls++
	return &testPrepared{resource: publicresource.Users}, nil
}

func (port *testIAMPort) Create(
	_ context.Context, scope platformtx.TenantScope,
	_ publicresource.ActorEvidence, _ publicresource.PreparedCreate,
) (publicresource.MutationResult, error) {
	port.scope = scope
	return port.fixture.mutation, nil
}

func (port *testIAMPort) Get(
	_ context.Context, scope platformtx.TenantScope,
	_ publicresource.ActorEvidence, _ publicresource.GetInput,
) (publicresource.ReadResult, error) {
	port.scope = scope
	return port.fixture.read, nil
}

func (port *testIAMPort) Query(
	_ context.Context, scope platformtx.TenantScope,
	_ publicresource.ActorEvidence, _ publicresource.QueryInput,
) (publicresource.QueryResult, error) {
	port.scope = scope
	return port.fixture.query, nil
}

func (port *testIAMPort) Update(
	_ context.Context, scope platformtx.TenantScope,
	_ publicresource.ActorEvidence, _ publicresource.UpdateInput,
) (publicresource.MutationResult, error) {
	port.scope = scope
	return port.fixture.mutation, nil
}

func (port *testIAMPort) Delete(
	_ context.Context, scope platformtx.TenantScope,
	_ publicresource.ActorEvidence, _ publicresource.DeleteInput,
) (publicresource.DeleteResult, error) {
	port.scope = scope
	return port.fixture.deleted, nil
}

func (port *testIAMPort) ChangeState(
	_ context.Context, scope platformtx.TenantScope,
	_ publicresource.ActorEvidence, _ publicresource.StateInput,
) (publicresource.MutationResult, error) {
	port.scope = scope
	return port.fixture.mutation, nil
}

type testResolver struct {
	document layoutapi.PresentationDocument
	err      error
	scope    platformtx.TenantScope
	calls    int
}

func (resolver *testResolver) ResolvePresentation(
	_ context.Context,
	scope platformtx.TenantScope,
	_ layoutapi.ResourceKind,
	_ layoutapi.PresentationFieldCatalog,
) (layoutapi.PresentationDocument, error) {
	resolver.calls++
	resolver.scope = scope
	return resolver.document, resolver.err
}

type testPrepared struct {
	resource  publicresource.Resource
	destroyed bool
}

func (prepared *testPrepared) Resource() (publicresource.Resource, error) {
	if prepared == nil || prepared.destroyed || !prepared.resource.Valid() {
		return "", publicresource.ErrContract
	}
	return prepared.resource, nil
}

func (*testPrepared) Contract() string { return publicresource.PreparedCreateContract }

func (prepared *testPrepared) Destroy() { prepared.destroyed = true }

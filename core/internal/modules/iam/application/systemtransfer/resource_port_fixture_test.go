package systemtransfer

import (
	"context"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type fixturePreparedCreate struct{ resource publicresource.Resource }

func (prepared *fixturePreparedCreate) Resource() (publicresource.Resource, error) {
	return prepared.resource, nil
}
func (*fixturePreparedCreate) Contract() string {
	return publicresource.PreparedCreateContract
}
func (*fixturePreparedCreate) Destroy() {}

type fixtureResourcePort struct {
	result           publicresource.MutationResult
	prepareCalls     int
	createCalls      int
	updateCalls      int
	createKey        string
	updateKey        string
	updateRevision   int64
	commandError     error
	preparedResource publicresource.Resource
}

func (port *fixtureResourcePort) Catalog(
	context.Context,
	platformtx.TenantScope,
	publicresource.ActorEvidence,
	publicresource.Resource,
) (publicresource.FieldCatalog, error) {
	return publicresource.FieldCatalog{}, publicresource.ErrUnavailable
}

func (port *fixtureResourcePort) PrepareCreate(
	_ context.Context,
	input publicresource.CreateInput,
) (publicresource.PreparedCreate, error) {
	port.prepareCalls++
	resource, err := input.Resource()
	if err != nil {
		return nil, err
	}
	port.preparedResource = resource
	port.createKey, _ = input.IdempotencyKey()
	return &fixturePreparedCreate{resource: resource}, nil
}

func (port *fixtureResourcePort) Create(
	context.Context,
	platformtx.TenantScope,
	publicresource.ActorEvidence,
	publicresource.PreparedCreate,
) (publicresource.MutationResult, error) {
	port.createCalls++
	return port.result, port.commandError
}

func (port *fixtureResourcePort) Get(
	context.Context,
	platformtx.TenantScope,
	publicresource.ActorEvidence,
	publicresource.GetInput,
) (publicresource.ReadResult, error) {
	return publicresource.ReadResult{}, publicresource.ErrUnavailable
}

func (port *fixtureResourcePort) Query(
	context.Context,
	platformtx.TenantScope,
	publicresource.ActorEvidence,
	publicresource.QueryInput,
) (publicresource.QueryResult, error) {
	return publicresource.QueryResult{}, publicresource.ErrUnavailable
}

func (port *fixtureResourcePort) Update(
	_ context.Context,
	_ platformtx.TenantScope,
	_ publicresource.ActorEvidence,
	input publicresource.UpdateInput,
) (publicresource.MutationResult, error) {
	port.updateCalls++
	_, port.updateRevision, _ = input.Identity()
	port.updateKey, _ = input.IdempotencyKey()
	return port.result, port.commandError
}

func (port *fixtureResourcePort) Delete(
	context.Context,
	platformtx.TenantScope,
	publicresource.ActorEvidence,
	publicresource.DeleteInput,
) (publicresource.DeleteResult, error) {
	return publicresource.DeleteResult{}, publicresource.ErrUnavailable
}

func (port *fixtureResourcePort) ChangeState(
	context.Context,
	platformtx.TenantScope,
	publicresource.ActorEvidence,
	publicresource.StateInput,
) (publicresource.MutationResult, error) {
	return publicresource.MutationResult{}, publicresource.ErrUnavailable
}

var _ publicresource.Port = (*fixtureResourcePort)(nil)

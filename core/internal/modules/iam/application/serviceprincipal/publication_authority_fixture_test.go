package serviceprincipal

import (
	"context"
	"sync"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type publicationAuthorityFixturePort struct {
	mu          sync.Mutex
	permission  uuid.UUID
	principal   PublicationAuthorityPrincipal
	grant       PublicationAuthorityGrant
	createCalls int
	grantCalls  int
}

func (port *publicationAuthorityFixturePort) ManagePermissionID(
	context.Context, platformtx.TenantScope,
) (uuid.UUID, error) {
	return port.permission, nil
}

func (port *publicationAuthorityFixturePort) PublicationPrincipal(
	context.Context, platformtx.TenantScope,
) (PublicationAuthorityPrincipal, error) {
	port.mu.Lock()
	defer port.mu.Unlock()
	return port.principal, nil
}

func (port *publicationAuthorityFixturePort) PublicationGrant(
	context.Context, platformtx.TenantScope, uuid.UUID,
) (PublicationAuthorityGrant, error) {
	port.mu.Lock()
	defer port.mu.Unlock()
	return port.grant, nil
}

func (port *publicationAuthorityFixturePort) Create(
	_ context.Context, _ platformtx.TenantScope, command CreateCommand,
) (MutationResult, error) {
	port.mu.Lock()
	defer port.mu.Unlock()
	port.createCalls++
	port.principal = PublicationAuthorityPrincipal{
		ID: command.ID, Found: true, Kind: "workflow", State: "active",
	}
	return MutationResult{
		ID: command.ID, Revision: 1, AuthorizationClock: 2,
	}, nil
}

func (port *publicationAuthorityFixturePort) Grant(
	_ context.Context, _ platformtx.TenantScope, command GrantCommand,
) (MutationResult, error) {
	port.mu.Lock()
	defer port.mu.Unlock()
	port.grantCalls++
	port.grant = PublicationAuthorityGrant{
		PublicationAuthority: PublicationAuthority{
			ServicePrincipalID: command.PrincipalID,
			GrantID:            command.GrantID, GrantRevision: 1,
		},
		Found: true, State: "active", Capability: command.Capability,
		Scope: command.Scope, ManifestSHA256: command.ManifestSHA256,
	}
	return MutationResult{
		ID: command.GrantID, Revision: 1, AuthorizationClock: 3,
	}, nil
}

func (*publicationAuthorityFixturePort) Disable(
	context.Context, platformtx.TenantScope, StateCommand,
) (MutationResult, error) {
	return MutationResult{}, ErrEvidenceMissing
}
func (*publicationAuthorityFixturePort) Enable(
	context.Context, platformtx.TenantScope, StateCommand,
) (MutationResult, error) {
	return MutationResult{}, ErrEvidenceMissing
}
func (*publicationAuthorityFixturePort) Retire(
	context.Context, platformtx.TenantScope, StateCommand,
) (MutationResult, error) {
	return MutationResult{}, ErrEvidenceMissing
}
func (*publicationAuthorityFixturePort) Revoke(
	context.Context, platformtx.TenantScope, RevokeCommand,
) (MutationResult, error) {
	return MutationResult{}, ErrEvidenceMissing
}

var _ CommandPort = (*publicationAuthorityFixturePort)(nil)
var _ PublicationAuthorityPort = (*publicationAuthorityFixturePort)(nil)

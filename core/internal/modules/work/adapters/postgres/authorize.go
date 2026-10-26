package postgres

import (
	"context"
	_ "embed"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
	"github.com/google/uuid"
)

//go:embed sql/tenant/authorize_mutation.sql
var authorizeMutationSQL string

type Authorizer struct {
	principals ports.PrincipalResolver
	access     *ownerAccess
}

func NewAuthorizer(principals ports.PrincipalResolver) (*Authorizer, error) {
	if principals == nil || authorizeMutationSQL == "" {
		return nil, application.ErrUnavailable
	}
	access, err := newOwnerAccess()
	if err != nil {
		return nil, application.ErrUnavailable
	}
	return &Authorizer{principals: principals, access: access}, nil
}

func (authorizer *Authorizer) Authorize(ctx context.Context, local ports.Scope,
	authority string, resource domain.ID,
) error {
	bridge, ok := local.(*scopeBridge)
	if authorizer == nil || !ok || !scopeUsable(ctx, bridge) || bridge.SnapshotQuery() {
		return application.ErrForbidden
	}
	snapshot, err := authorizer.principals.Resolve(ctx, local)
	if err != nil || snapshot.PermissionRevision < 0 {
		return application.ErrForbidden
	}
	kinds, ids := principalArrays(snapshot.Principals)
	database, err := authorizer.access.bind(ctx, bridge)
	if err != nil {
		return application.ErrUnavailable
	}
	var allowed bool
	err = database.QueryRow(ctx, authorizeMutationSQL, authority, uuid.UUID(resource),
		bridge.scope.ActorID(), kinds, ids).Scan(&allowed)
	if err != nil || !allowed {
		return application.ErrForbidden
	}
	return nil
}

func principalArrays(principals []domain.Principal) ([]string, []uuid.UUID) {
	kinds := make([]string, 0, len(principals))
	ids := make([]uuid.UUID, 0, len(principals))
	for _, principal := range principals {
		kinds, ids = append(kinds, string(principal.Kind)), append(ids, uuid.UUID(principal.ID))
	}
	return kinds, ids
}

var _ ports.Authorizer = (*Authorizer)(nil)

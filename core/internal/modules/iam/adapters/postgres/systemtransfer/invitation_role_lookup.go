package systemtransfer

import (
	"context"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	invitationapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/invitation"
	systemtransferapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
	"github.com/google/uuid"
)

func (store *WorkerStore) FindInvitationRole(ctx context.Context,
	scope tenantScope, evidence publicresource.ActorEvidence,
	roleID publicresource.ID,
) (invitationapp.RoleScope, error) {
	tenantID, tenantErr := evidence.TenantID()
	actorID, actorErr := evidence.ActorID()
	if tenantErr != nil || actorErr != nil || !roleID.Valid() ||
		uuid.UUID(tenantID) != scope.TenantID() ||
		uuid.UUID(actorID) != scope.ActorID() {
		return invitationapp.RoleScope{}, systemtransferapp.ErrInvitationDenied
	}
	bound, err := store.bind(ctx, scope)
	if err != nil {
		return invitationapp.RoleScope{}, systemtransferapp.ErrInvitationUnavailable
	}
	var id, scopeID uuid.UUID
	var kind string
	err = bound.database.QueryRow(ctx, resolveImportRoleSQL,
		uuid.UUID(roleID), scope.TenantID()).Scan(&id, &kind, &scopeID)
	if bound.notFound != nil && bound.notFound(err) {
		return invitationapp.RoleScope{}, systemtransferapp.ErrInvitationDenied
	}
	roleScope := invitationapp.RoleScope{
		Kind: invitationapp.RoleScopeKind(kind), ID: scopeID,
	}
	if roleScope.Kind == invitationapp.RoleScopeOrganization {
		roleScope.OrganizationID = scopeID
	}
	if err != nil || id != uuid.UUID(roleID) || roleScope.Validate() != nil {
		return invitationapp.RoleScope{}, systemtransferapp.ErrInvitationUnavailable
	}
	return roleScope, nil
}

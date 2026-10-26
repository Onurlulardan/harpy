package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/invitation"
	"github.com/google/uuid"
)

func (store *invitationStore) LookupRole(
	ctx context.Context,
	scope tenantScope,
	lookup invitation.RoleLookup,
) (invitation.RoleTarget, error) {
	if !invitationSnapshotScopeUsable(ctx, scope) {
		return invitation.RoleTarget{}, invitation.ErrInvalid
	}
	bound, err := store.bindTenant(scope)
	if err != nil {
		return invitation.RoleTarget{}, err
	}
	var target invitation.RoleTarget
	var kind string
	var organizationID *uuid.UUID
	err = bound.database.QueryRow(
		ctx, invitationRoleLookupSQL, scope.ActorID(), lookup.Name,
	).Scan(&target.ID, &target.Name, &kind, &target.Scope.ID, &organizationID)
	if err != nil {
		return invitation.RoleTarget{}, mapInvitationError(err)
	}
	target.Scope.Kind = invitation.RoleScopeKind(kind)
	if organizationID != nil {
		target.Scope.OrganizationID = *organizationID
	}
	if target.Validate() != nil {
		return invitation.RoleTarget{}, invitation.ErrUnavailable
	}
	return target, nil
}

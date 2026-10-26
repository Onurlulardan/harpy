package iam

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/invitation"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *invitationHTTPApplication) invitationRole(
	ctx context.Context, claims accesstoken.Claims, name string,
) (invitation.RoleTarget, error) {
	request, err := application.snapshotScope(ctx, claims)
	if err != nil {
		return invitation.RoleTarget{}, err
	}
	var result invitation.RoleTarget
	err = application.coordinator.InTenant(
		ctx, request, func(scoped context.Context, scope platformtx.TenantScope) error {
			var lookupErr error
			result, lookupErr = application.runtime.service.LookupRole(
				scoped, scope, invitation.RoleLookup{Name: name},
			)
			return lookupErr
		},
	)
	if err != nil {
		return invitation.RoleTarget{}, mapInvitationApplicationError(err)
	}
	return result, nil
}

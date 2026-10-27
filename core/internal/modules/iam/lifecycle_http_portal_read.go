package iam

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycle"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *lifecycleHTTPApplication) ReadPortalSetup(
	ctx context.Context, claims accesstoken.Claims, bindingID uuid.UUID,
) (lifecycle.PortalSetup, error) {
	if bindingID == uuid.Nil {
		return lifecycle.PortalSetup{}, ErrLifecycleHTTPInvalid
	}
	var result lifecycle.PortalSetup
	err := application.inSnapshot(ctx, claims, "read",
		func(scoped context.Context, scope platformtx.TenantScope, permissionID uuid.UUID) error {
			var readErr error
			result, readErr = application.runtime.service.ResolvePortalSetup(
				scoped, scope, lifecycle.PortalSetupQuery{
					BindingID: bindingID, ActorID: claims.Subject(),
					PermissionID:                 permissionID,
					ExpectedTokenVersion:         claims.TokenVersion(),
					ExpectedAuthorizationVersion: claims.AuthorizationVersion(),
				},
			)
			return readErr
		})
	return result, err
}

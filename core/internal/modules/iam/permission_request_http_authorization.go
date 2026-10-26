package iam

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authorization"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *permissionRequestHTTPApplication) authorize(
	ctx context.Context,
	scope platformtx.TenantScope,
	evidence identityEvidenceSource,
	action string,
) error {
	request, err := authorization.NewRequest(
		"iam.permission_requests", action,
		evidence.TokenVersion(), evidence.AuthorizationVersion(),
	)
	if err != nil {
		return ErrPermissionRequestHTTPForbidden
	}
	decision, err := application.authorizer.Authorize(ctx, scope, request)
	if err != nil || decision.Validate() != nil {
		return ErrPermissionRequestHTTPUnavailable
	}
	if !decision.Allowed() {
		return ErrPermissionRequestHTTPForbidden
	}
	return nil
}

package iam

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
)

func (runtime *WorkflowAccessRuntime) Verify(
	ctx context.Context,
	request WorkflowAccessRequest,
) (WorkflowAccessDecision, error) {
	if runtime == nil || runtime.shared == nil || runtime.shared.security == nil ||
		runtime.shared.clock == nil || ctx == nil || ctx.Err() != nil {
		return WorkflowAccessDecision{}, ErrWorkflowAccessUnavailable
	}
	target, targetValid := resolveWorkflowAccess(request.Permission, request.Operation)
	wire, bearerValid := formsRecordBearer(request.AuthorizationValues)
	now := runtime.shared.clock().UTC().Truncate(time.Second)
	if !targetValid || !bearerValid || now.IsZero() || now.Unix() <= 0 {
		return WorkflowAccessDecision{}, ErrWorkflowAccessAuthentication
	}
	claims, err := runtime.shared.security.accessVerifier.Verify(ctx, wire, now)
	if err != nil || claims.AccessMode() != accesstoken.AccessModeOperator ||
		!identityHTTPClientAllowed(claims.Client()) ||
		!runtime.tenantHeadersMatch(ctx, request, claims.Tenant()) {
		return WorkflowAccessDecision{}, ErrWorkflowAccessAuthentication
	}
	if _, portal := claims.PortalBinding(); portal {
		return WorkflowAccessDecision{}, ErrWorkflowAccessForbidden
	}
	authorizationDecision, err := runtime.authorize(ctx, claims, target)
	if err != nil {
		return WorkflowAccessDecision{}, err
	}
	if request.Operation == "workflow.publication_authority.provision" {
		if _, err = runtime.authorize(
			ctx, claims, workflowPublicationAuthorityManageTarget,
		); err != nil {
			return WorkflowAccessDecision{}, err
		}
		purpose, authenticatedAt, password, webAuthn, totp, valid :=
			workflowPublicationAuthorityStepUp(claims, request.TargetID, now)
		if !valid {
			return WorkflowAccessDecision{}, ErrWorkflowAccessForbidden
		}
		decision, decisionErr := NewWorkflowPublicationAuthorityAccessDecision(
			claims.Subject(), claims.Tenant(), claims.TokenVersion(),
			claims.AuthorizationVersion(), authorizationDecision.AuthorizationClock(),
			purpose, authenticatedAt, password, webAuthn, totp,
		)
		if decisionErr != nil {
			return WorkflowAccessDecision{}, ErrWorkflowAccessUnavailable
		}
		return decision, nil
	}
	decision, err := NewWorkflowAccessDecision(claims.Subject(), claims.Tenant(),
		claims.TokenVersion(), claims.AuthorizationVersion(),
		authorizationDecision.AuthorizationClock())
	if err != nil {
		return WorkflowAccessDecision{}, ErrWorkflowAccessUnavailable
	}
	return decision, nil
}

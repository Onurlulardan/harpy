package api

import (
	"context"
	"errors"
	"fmt"

	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
	usagemodule "github.com/Onurlulardan/harpy/core/internal/modules/usage"
)

type usageIAMAccess interface {
	Verify(context.Context, iammodule.UsageAccessRequest) (
		iammodule.UsageAccessDecision, error,
	)
}

type usageAccessVerifier struct{ iam usageIAMAccess }

func newUsageAccessVerifier(iam usageIAMAccess) (*usageAccessVerifier, error) {
	if iam == nil {
		return nil, fmt.Errorf("Usage IAM access is required")
	}
	return &usageAccessVerifier{iam: iam}, nil
}

func (verifier *usageAccessVerifier) Verify(
	ctx context.Context, request usagemodule.HTTPAccessRequest,
) (usagemodule.Access, error) {
	if verifier == nil || verifier.iam == nil {
		return usagemodule.Access{}, usagemodule.ErrHTTPUnavailable
	}
	decision, err := verifier.iam.Verify(ctx, iammodule.UsageAccessRequest{
		AuthorizationValues: request.AuthorizationValues,
		TenantIDValues:      request.TenantIDValues, TenantSlugValues: request.TenantSlugValues,
		Scope: string(request.Scope), Permission: request.Permission, Operation: request.Operation,
	})
	if err != nil {
		switch {
		case errors.Is(err, iammodule.ErrUsageAccessAuthentication):
			return usagemodule.Access{}, usagemodule.ErrHTTPAuthentication
		case errors.Is(err, iammodule.ErrUsageAccessForbidden):
			return usagemodule.Access{}, usagemodule.ErrHTTPForbidden
		default:
			return usagemodule.Access{}, usagemodule.ErrHTTPUnavailable
		}
	}
	access, err := usagemodule.NewAccess(usagemodule.AccessInput{
		ActorID: decision.ActorID(), SubjectTenantID: decision.SubjectTenantID(),
		TargetTenantID: decision.TargetTenantID(), EnvironmentID: decision.EnvironmentID(),
		TokenVersion: decision.TokenVersion(), AuthorizationVersion: decision.AuthorizationVersion(),
		AuthorityClock: decision.AuthorityClock(), Scope: request.Scope,
		Mode:           usagemodule.DeploymentMode(decision.Mode()),
		Purpose:        usagemodule.TenantPurpose(decision.Purpose()),
		ControlPrimary: decision.ControlPrimary(), Permission: request.Permission,
		Operation: request.Operation,
	})
	if err != nil {
		return usagemodule.Access{}, usagemodule.ErrHTTPUnavailable
	}
	return access, nil
}

var _ usagemodule.HTTPAccessVerifier = (*usageAccessVerifier)(nil)

package api

import (
	"context"
	"errors"
	"fmt"

	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
	tenancymodule "github.com/Onurlulardan/harpy/core/internal/modules/tenancy"
)

type tenantSettingsIAMAccess interface {
	Verify(context.Context, iammodule.TenantSettingsAccessRequest) (
		iammodule.TenantSettingsAccessDecision, error,
	)
}

type tenantSettingsAccessVerifier struct{ iam tenantSettingsIAMAccess }

func newTenantSettingsAccessVerifier(
	iam tenantSettingsIAMAccess,
) (*tenantSettingsAccessVerifier, error) {
	if iam == nil {
		return nil, fmt.Errorf("tenant settings IAM access is required")
	}
	return &tenantSettingsAccessVerifier{iam: iam}, nil
}

func (verifier *tenantSettingsAccessVerifier) Verify(
	ctx context.Context, request tenancymodule.TenantSettingsAccessRequest,
) (tenancymodule.TenantSettingsAccess, error) {
	if verifier == nil || verifier.iam == nil {
		return tenancymodule.TenantSettingsAccess{},
			tenancymodule.ErrTenantSettingsUnavailable
	}
	decision, err := verifier.iam.Verify(ctx, iammodule.TenantSettingsAccessRequest{
		AuthorizationValues: request.AuthorizationValues,
		TenantIDValues:      request.TenantIDValues,
		TenantSlugValues:    request.TenantSlugValues,
		Permission:          string(request.Permission),
		Operation:           request.Operation,
	})
	if err != nil {
		switch {
		case errors.Is(err, iammodule.ErrTenantSettingsAccessAuthentication):
			return tenancymodule.TenantSettingsAccess{},
				tenancymodule.ErrTenantSettingsAuthentication
		case errors.Is(err, iammodule.ErrTenantSettingsAccessForbidden):
			return tenancymodule.TenantSettingsAccess{},
				tenancymodule.ErrTenantSettingsForbidden
		default:
			return tenancymodule.TenantSettingsAccess{},
				tenancymodule.ErrTenantSettingsUnavailable
		}
	}
	access, err := tenancymodule.NewTenantSettingsAccess(
		decision.PrincipalID(), decision.TenantID(), decision.TokenVersion(),
		decision.AuthorizationVersion(), decision.AuthorizationClock(),
		request.Permission, request.Operation,
	)
	if err != nil {
		return tenancymodule.TenantSettingsAccess{},
			tenancymodule.ErrTenantSettingsUnavailable
	}
	return access, nil
}

var _ tenancymodule.TenantSettingsAccessVerifier = (*tenantSettingsAccessVerifier)(nil)

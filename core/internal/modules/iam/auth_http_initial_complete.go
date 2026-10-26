package iam

import (
	"context"
	"errors"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authentication"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/password"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *authHTTPApplication) CompleteInitialPassword(
	ctx context.Context,
	input AuthHTTPInitialPasswordInput,
	presented *authentication.PresentedInitialPassword,
) (AuthHTTPGrant, error) {
	defer clear(input.NewPassword)
	if presented != nil {
		defer presented.Destroy()
	}
	if application == nil || ctx == nil || ctx.Err() != nil ||
		presented == nil || !validLoginText(input.TenantSlug, 128) {
		return nil, ErrAuthHTTPInvalidCredentials
	}
	if _, valid := authHTTPPeerIP(ctx); !valid {
		return nil, ErrAuthHTTPUnavailable
	}
	selector, err := presented.Selector()
	if err != nil {
		return nil, ErrAuthHTTPInvalidCredentials
	}
	tenantID, err := application.pool.ResolveActiveTenantIDBySlug(ctx, input.TenantSlug)
	if err != nil {
		return nil, ErrAuthHTTPInvalidCredentials
	}
	binding := authentication.InitialPasswordRouteBinding(tenantID, selector)
	lookup, err := platformtx.NewTenantRouteCredentialLookup(
		"iam.authentication.initial.complete",
		platformtx.CredentialInitialPassword, tenantID, binding,
	)
	clear(binding[:])
	if err != nil {
		return nil, ErrAuthHTTPUnavailable
	}
	var grant authentication.Grant
	var policyOutcome error
	err = application.coordinator.InCredentialTenant(
		ctx, platformtx.CredentialScopeRequest{Lookup: lookup, Resolver: application.resolver},
		func(scoped context.Context, scope platformtx.CredentialTenantScope) error {
			verified, verifyErr := application.initial.Resolve(scoped, scope, presented)
			if verifyErr != nil {
				return verifyErr
			}
			defer verified.Destroy()
			secret, policyErr := application.passwords.NewSecret(
				scoped, input.NewPassword, []string{input.TenantSlug},
			)
			if policyErr != nil {
				if errors.Is(policyErr, password.ErrPasswordPolicyUnavailable) {
					return policyErr
				}
				rejectErr := application.initial.Reject(scoped, scope, &verified)
				if !errors.Is(rejectErr, authentication.ErrInitialPasswordPolicy) {
					return rejectErr
				}
				policyOutcome = authentication.ErrInitialPasswordPolicy
				return nil
			}
			defer secret.Destroy()
			var completeErr error
			grant, completeErr = application.initial.Finalize(
				scoped, scope, &verified, secret, application.deviceEvidence(scoped),
			)
			return completeErr
		},
	)
	if err != nil {
		grant.Destroy()
		return nil, mapHTTPInitialPasswordError(err)
	}
	if policyOutcome != nil {
		return nil, mapHTTPInitialPasswordError(policyOutcome)
	}
	return &grant, nil
}

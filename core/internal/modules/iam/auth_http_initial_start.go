package iam

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authentication"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/password"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *authHTTPApplication) StartInitialPassword(
	ctx context.Context,
	input AuthHTTPLoginInput,
) (AuthHTTPInitialChallenge, error) {
	defer clear(input.Password)
	if application == nil || ctx == nil || ctx.Err() != nil ||
		input.Mode != AuthHTTPLoginOperator {
		return nil, ErrAuthHTTPUnavailable
	}
	peerIP, valid := authHTTPPeerIP(ctx)
	if !valid {
		return nil, ErrAuthHTTPUnavailable
	}
	value, kind, _, err := normalizeHTTPLogin(input)
	if err != nil {
		return nil, ErrAuthHTTPInvalidCredentials
	}
	environment, err := application.loginEnvironmentBuckets(
		input, value.Normalized(), peerIP,
	)
	if err != nil {
		return nil, err
	}
	if err := application.throttleAnyBlocked(ctx, environment); err != nil {
		return nil, err
	}
	tenantID, err := application.pool.ResolveActiveTenantIDBySlug(ctx, input.TenantSlug)
	if err != nil {
		return nil, ErrAuthHTTPInvalidCredentials
	}
	keyID, key, err := application.directory.Active()
	if err != nil {
		return nil, ErrAuthHTTPUnavailable
	}
	lookupHMAC := loginDirectoryHMAC(key, value.Normalized())
	clear(key[:])
	lookup, err := platformtx.NewInitialPasswordCredentialLookup(
		"iam.authentication.initial.start", tenantID, kind, keyID, lookupHMAC,
	)
	clear(lookupHMAC[:])
	if err != nil {
		return nil, ErrAuthHTTPUnavailable
	}
	secret, err := password.NewVerificationSecret(input.Password)
	if err != nil {
		return nil, ErrAuthHTTPInvalidCredentials
	}
	defer secret.Destroy()
	var challenge authentication.InitialPasswordChallenge
	err = application.coordinator.InCredentialTenant(
		ctx, platformtx.CredentialScopeRequest{Lookup: lookup, Resolver: application.resolver},
		func(scoped context.Context, scope platformtx.CredentialTenantScope) error {
			buckets, bucketErr := application.loginTenantBuckets(
				value.Normalized(), peerIP, tenantID,
				scope.AuthorityGroupGeneration(), scope.IAMGeneration(),
			)
			if bucketErr != nil {
				return bucketErr
			}
			if blockErr := application.throttleAnyBlocked(scoped, buckets); blockErr != nil {
				return blockErr
			}
			var startErr error
			challenge, startErr = application.initial.Claim(scoped, scope, secret)
			return startErr
		},
	)
	if err != nil {
		challenge.Destroy()
		return nil, mapHTTPInitialPasswordError(err)
	}
	return &challenge, nil
}

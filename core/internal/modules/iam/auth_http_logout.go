package iam

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authentication"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/refreshtoken"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *authHTTPApplication) Logout(
	ctx context.Context, presented *refreshtoken.Presented,
) error {
	if application == nil || ctx == nil || ctx.Err() != nil || presented == nil {
		return ErrAuthHTTPInvalidSession
	}
	defer presented.Destroy()
	peerIP, valid := authHTTPPeerIP(ctx)
	if !valid {
		return ErrAuthHTTPUnavailable
	}
	lookup, err := refreshCredentialLookup(
		"iam.authentication.revoke", presented,
	)
	if err != nil {
		return ErrAuthHTTPInvalidSession
	}
	environmentBuckets, err := application.refreshEnvironmentBuckets(lookup, peerIP)
	if err != nil {
		return err
	}
	if err := application.throttleAnyBlocked(ctx, environmentBuckets); err != nil {
		return mapHTTPSessionError(err)
	}
	var tenantBuckets []authHTTPThrottleBucket
	err = application.coordinator.InCredentialTenant(
		ctx, platformtx.CredentialScopeRequest{Lookup: lookup, Resolver: application.resolver},
		func(scoped context.Context, scope platformtx.CredentialTenantScope) error {
			tenantBuckets, err = application.refreshTenantBuckets(scope, peerIP)
			if err != nil {
				return err
			}
			if err := application.throttleAnyBlocked(scoped, tenantBuckets); err != nil {
				return err
			}
			return application.authenticator.RevokeSession(
				scoped, scope, authentication.LogoutRequest{Presented: presented},
			)
		},
	)
	if err != nil {
		if mapHTTPAuthenticationError(err) == ErrAuthHTTPInvalidCredentials {
			buckets := allAuthHTTPThrottleBuckets(environmentBuckets, tenantBuckets)
			if recordErr := application.recordThrottleFailure(ctx, buckets...); recordErr != nil {
				return recordErr
			}
		}
		return mapHTTPSessionError(err)
	}
	return nil
}

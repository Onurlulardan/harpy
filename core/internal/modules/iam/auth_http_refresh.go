package iam

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authentication"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/refreshtoken"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *authHTTPApplication) Refresh(
	ctx context.Context,
	presented *refreshtoken.Presented,
) (AuthHTTPGrant, error) {
	if application == nil || ctx == nil || ctx.Err() != nil || presented == nil {
		return nil, ErrAuthHTTPInvalidSession
	}
	peerIP, valid := authHTTPPeerIP(ctx)
	if !valid {
		return nil, ErrAuthHTTPUnavailable
	}
	defer presented.Destroy()
	lookup, err := refreshCredentialLookup(
		"iam.authentication.rotate", presented,
	)
	if err != nil {
		return nil, ErrAuthHTTPInvalidSession
	}
	environmentBuckets, err := application.refreshEnvironmentBuckets(lookup, peerIP)
	if err != nil {
		return nil, err
	}
	if err := application.throttleAnyBlocked(ctx, environmentBuckets); err != nil {
		return nil, mapHTTPSessionError(err)
	}
	var rotation authentication.Rotation
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
			var rotateErr error
			rotation, rotateErr = application.authenticator.RotateBrowser(
				scoped, scope, presented,
			)
			return rotateErr
		},
	)
	if err != nil {
		rotation.Destroy()
		if mapHTTPAuthenticationError(err) == ErrAuthHTTPInvalidCredentials {
			buckets := allAuthHTTPThrottleBuckets(environmentBuckets, tenantBuckets)
			if recordErr := application.recordThrottleFailure(ctx, buckets...); recordErr != nil {
				return nil, recordErr
			}
		}
		return nil, mapHTTPSessionError(err)
	}
	if rotation.Outcome() != authentication.RefreshRotatedOutcome {
		rotation.Destroy()
		buckets := allAuthHTTPThrottleBuckets(environmentBuckets, tenantBuckets)
		if err := application.recordThrottleFailure(ctx, buckets...); err != nil {
			return nil, err
		}
		return nil, ErrAuthHTTPInvalidSession
	}
	return &rotation, nil
}

func refreshCredentialLookup(
	operationID string,
	presented *refreshtoken.Presented,
) (platformtx.CredentialLookup, error) {
	selector, err := presented.Selector()
	if err != nil {
		return platformtx.CredentialLookup{}, err
	}
	return platformtx.NewRefreshCredentialLookup(operationID, selector)
}

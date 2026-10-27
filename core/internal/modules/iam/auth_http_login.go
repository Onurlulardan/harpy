package iam

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authentication"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/password"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *authHTTPApplication) Login(
	ctx context.Context, input AuthHTTPLoginInput,
) (AuthHTTPGrant, error) {
	defer clear(input.Password)
	if application == nil || ctx == nil || ctx.Err() != nil {
		return nil, ErrAuthHTTPUnavailable
	}
	peerIP, valid := authHTTPPeerIP(ctx)
	if !valid {
		if equalizeErr := application.equalizeLogin(input.Password); equalizeErr != nil {
			return nil, equalizeErr
		}
		return nil, ErrAuthHTTPUnavailable
	}
	value, kind, client, err := normalizeHTTPLogin(input)
	if err != nil {
		if equalizeErr := application.equalizeLogin(input.Password); equalizeErr != nil {
			return nil, equalizeErr
		}
		return nil, ErrAuthHTTPInvalidCredentials
	}
	environmentBuckets, err := application.loginEnvironmentBuckets(
		input, value.Normalized(), peerIP,
	)
	if err != nil {
		return nil, err
	}
	if err := application.throttleAnyBlocked(ctx, environmentBuckets); err != nil {
		if equalizeErr := application.equalizeLogin(input.Password); equalizeErr != nil {
			return nil, equalizeErr
		}
		return nil, err
	}
	tenantID, err := application.pool.ResolveActiveTenantIDBySlug(ctx, input.TenantSlug)
	if err != nil {
		if recordErr := application.recordThrottleFailure(
			ctx, environmentBuckets...,
		); recordErr != nil {
			return nil, recordErr
		}
		if equalizeErr := application.equalizeLogin(input.Password); equalizeErr != nil {
			return nil, equalizeErr
		}
		return nil, ErrAuthHTTPInvalidCredentials
	}
	keyID, key, err := application.directory.Active()
	if err != nil {
		return nil, ErrAuthHTTPUnavailable
	}
	lookupHMAC := loginDirectoryHMAC(key, value.Normalized())
	clear(key[:])
	lookup, err := platformtx.NewLoginCredentialLookup(
		"iam.authentication.issue", tenantID, kind, keyID, lookupHMAC,
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
	var grant authentication.Grant
	var tenantBuckets []authHTTPThrottleBucket
	credentialAttempted := false
	err = application.coordinator.InCredentialTenant(
		ctx, platformtx.CredentialScopeRequest{Lookup: lookup, Resolver: application.resolver},
		func(scoped context.Context, scope platformtx.CredentialTenantScope) error {
			tenantBuckets, err = application.loginTenantBuckets(
				value.Normalized(), peerIP, tenantID,
				scope.AuthorityGroupGeneration(), scope.IAMGeneration(),
			)
			if err != nil {
				return err
			}
			if err := application.throttleAnyBlocked(scoped, tenantBuckets); err != nil {
				return err
			}
			credentialAttempted = true
			var issueErr error
			grant, issueErr = application.authenticator.Issue(
				scoped, scope, authentication.IssueRequest{
					Secret: secret, Client: client, Device: application.deviceEvidence(scoped),
				},
			)
			return issueErr
		},
	)
	if err != nil {
		grant.Destroy()
		if !credentialAttempted {
			if equalizeErr := application.equalizeLogin(input.Password); equalizeErr != nil {
				return nil, equalizeErr
			}
		}
		if mapHTTPAuthenticationError(err) == ErrAuthHTTPInvalidCredentials {
			buckets := append([]authHTTPThrottleBucket(nil), environmentBuckets...)
			buckets = append(buckets, tenantBuckets...)
			if recordErr := application.recordThrottleFailure(ctx, buckets...); recordErr != nil {
				return nil, recordErr
			}
		}
		return nil, mapHTTPAuthenticationError(err)
	}
	return &grant, nil
}

func (application *authHTTPApplication) loginEnvironmentBuckets(
	input AuthHTTPLoginInput, normalized string, peerIP string,
) ([]authHTTPThrottleBucket, error) {
	return application.throttleBuckets(
		"environment", "login", uuid.Nil, 0, 0,
		loginEnvironmentThrottleDimensions(input.TenantSlug, normalized, peerIP),
	)
}

func (application *authHTTPApplication) loginTenantBuckets(
	normalized, peerIP string,
	tenantID uuid.UUID,
	groupGeneration, iamGeneration int64,
) ([]authHTTPThrottleBucket, error) {
	return application.throttleBuckets(
		"tenant", "login", tenantID, groupGeneration, iamGeneration,
		loginTenantThrottleDimensions(normalized, peerIP),
	)
}

func loginDirectoryHMAC(key [32]byte, normalized string) [32]byte {
	mac := hmac.New(sha256.New, key[:])
	_, _ = mac.Write([]byte(normalized))
	var result [32]byte
	copy(result[:], mac.Sum(nil))
	return result
}

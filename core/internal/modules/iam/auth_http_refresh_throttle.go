package iam

import (
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *authHTTPApplication) refreshEnvironmentBuckets(
	lookup platformtx.CredentialLookup,
	peerIP string,
) ([]authHTTPThrottleBucket, error) {
	selector := lookup.Key()
	defer clear(selector)
	return application.throttleBuckets(
		"environment", "refresh", uuid.Nil, 0, 0,
		refreshThrottleDimensions(string(selector), peerIP),
	)
}

func (application *authHTTPApplication) refreshTenantBuckets(
	scope platformtx.CredentialTenantScope,
	peerIP string,
) ([]authHTTPThrottleBucket, error) {
	selector := scope.SourceSelector()
	return application.throttleBuckets(
		"tenant", "refresh", scope.TenantID(),
		scope.AuthorityGroupGeneration(), scope.IAMGeneration(),
		refreshThrottleDimensions(string(selector[:]), peerIP),
	)
}

func allAuthHTTPThrottleBuckets(
	environment, tenant []authHTTPThrottleBucket,
) []authHTTPThrottleBucket {
	result := make([]authHTTPThrottleBucket, 0, len(environment)+len(tenant))
	result = append(result, environment...)
	return append(result, tenant...)
}

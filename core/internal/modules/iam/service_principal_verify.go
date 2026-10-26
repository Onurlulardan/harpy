package iam

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *ServicePrincipalRuntime) Verify(
	ctx context.Context,
	scope platformtx.TenantScope,
	request ServicePrincipalPinnedRequest,
) error {
	if !runtime.Valid() {
		return ErrServicePrincipalRuntime
	}
	return runtime.verifier.Verify(ctx, scope, request)
}

func (runtime *ServicePrincipalRuntime) VerifyEvidence(
	ctx context.Context,
	scope platformtx.TenantScope,
	request ServicePrincipalPinnedRequest,
) (ServicePrincipalVerificationEvidence, error) {
	if !runtime.Valid() {
		return ServicePrincipalVerificationEvidence{}, ErrServicePrincipalRuntime
	}
	return runtime.verifier.VerifyEvidence(ctx, scope, request)
}

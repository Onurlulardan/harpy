package authentication

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (authenticator InitialPasswordAuthenticator) Reject(
	ctx context.Context,
	scope platformtx.CredentialTenantScope,
	verified *VerifiedInitialPassword,
) error {
	if !authenticator.valid || !verifiedInitialUsable(ctx, scope, verified) {
		return ErrAuthenticationContract
	}
	defer verified.destroy()
	continuation := verified.continuation
	result, err := authenticator.port.RejectInitialPassword(ctx, scope, InitialPasswordReject{
		Selector: continuation.Selector, SecretHMAC: verified.secretHMAC,
		CSRFHMAC: verified.csrfHMAC, ExpectedRevision: continuation.Revision,
	})
	if err != nil || result.PolicyAttempts != continuation.PolicyAttempts+1 ||
		result.Revision != continuation.Revision+1 ||
		(result.PolicyAttempts < InitialPasswordMaxAttempts && result.State != "active") ||
		(result.PolicyAttempts == InitialPasswordMaxAttempts && result.State != "exhausted") {
		return ErrAuthenticationStorage
	}
	return ErrInitialPasswordPolicy
}

func verifiedInitialUsable(
	ctx context.Context,
	scope platformtx.CredentialTenantScope,
	verified *VerifiedInitialPassword,
) bool {
	return verified != nil && verified.valid &&
		validCredentialScope(ctx, scope, platformtx.CredentialInitialPassword,
			platformtx.CredentialTenantRoute) &&
		verified.continuation.TenantID == scope.TenantID()
}

func (verified *VerifiedInitialPassword) destroy() {
	if verified == nil {
		return
	}
	clear(verified.secretHMAC[:])
	clear(verified.csrfHMAC[:])
	verified.continuation = InitialPasswordContinuation{}
	verified.valid = false
}

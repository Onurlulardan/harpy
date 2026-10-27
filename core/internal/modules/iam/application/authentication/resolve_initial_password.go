package authentication

import (
	"context"
	"crypto/subtle"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (authenticator InitialPasswordAuthenticator) Resolve(
	ctx context.Context,
	scope platformtx.CredentialTenantScope,
	presented *PresentedInitialPassword,
) (VerifiedInitialPassword, error) {
	if presented != nil {
		defer presented.Destroy()
	}
	if !authenticator.valid || presented == nil || !presented.active ||
		!validCredentialScope(ctx, scope, platformtx.CredentialInitialPassword,
			platformtx.CredentialTenantRoute) {
		return VerifiedInitialPassword{}, ErrAuthenticationContract
	}
	selector := presented.selector
	expectedRoute := InitialPasswordRouteBinding(scope.TenantID(), selector)
	actualRoute := scope.LocatorBindingSHA256()
	if subtle.ConstantTimeCompare(expectedRoute[:], actualRoute[:]) != 1 {
		return VerifiedInitialPassword{}, ErrInitialPasswordState
	}
	continuation, err := authenticator.port.InitialPasswordContinuation(ctx, scope, selector)
	now := authenticator.base.clock().UTC()
	if err != nil || !continuation.valid() || continuation.TenantID != scope.TenantID() ||
		continuation.Selector != selector || now.Before(continuation.IssuedAt) ||
		!now.Before(continuation.ExpiresAt) {
		return VerifiedInitialPassword{}, ErrInitialPasswordState
	}
	secretKey, err := authenticator.continuation.HMACKey(continuation.SecretKeyID)
	if err != nil || secretKey == [32]byte{} {
		return VerifiedInitialPassword{}, ErrAuthenticationCrypto
	}
	csrfKey, err := authenticator.csrf.HMACKey(continuation.CSRFKeyID)
	if err != nil || csrfKey == [32]byte{} {
		clear(secretKey[:])
		return VerifiedInitialPassword{}, ErrAuthenticationCrypto
	}
	matched := presented.matches(continuation.SecretHMAC, continuation.CSRFHMAC,
		secretKey, csrfKey, continuation.binding())
	clear(secretKey[:])
	clear(csrfKey[:])
	if !matched {
		return VerifiedInitialPassword{}, ErrInitialPasswordState
	}
	return VerifiedInitialPassword{continuation: continuation,
		secretHMAC: continuation.SecretHMAC, csrfHMAC: continuation.CSRFHMAC,
		valid: true}, nil
}

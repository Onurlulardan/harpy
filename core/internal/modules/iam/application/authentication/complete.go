package authentication

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (authenticator Authenticator) CompleteReauthentication(
	ctx context.Context,
	scope platformtx.CredentialTenantScope,
	request ReauthenticateRequest,
) (Grant, error) {
	defer request.destroy()
	if !authenticator.valid || request.validate() != nil ||
		!validCredentialScope(
			ctx, scope, platformtx.CredentialReauthentication,
			platformtx.CredentialSessionLocator,
		) {
		return Grant{}, ErrAuthenticationContract
	}
	lookup, err := request.lookup(scope)
	if err != nil {
		return Grant{}, err
	}
	credential, err := authenticator.port.ReauthenticationCredential(
		ctx, scope, lookup,
	)
	if err != nil {
		return Grant{}, ErrInvalidCredentials
	}
	now := authenticator.clock().UTC()
	if now.IsZero() {
		return Grant{}, ErrAuthenticationCrypto
	}
	if !now.Before(credential.ExpiresAt) {
		_ = authenticator.rejectReauthentication(ctx, scope, credential, false, true)
		return Grant{}, ErrInvalidCredentials
	}
	if credential.validate(lookup, now) != nil {
		return Grant{}, ErrInvalidCredentials
	}
	proof, quarantine, err := authenticator.verifyReauthenticationProof(
		ctx, request, credential, now,
	)
	if err != nil {
		if proofFailure(err) {
			if rejectErr := authenticator.rejectReauthentication(
				ctx, scope, credential, quarantine, false,
			); rejectErr != nil {
				return Grant{}, rejectErr
			}
			return Grant{}, ErrInvalidCredentials
		}
		return Grant{}, err
	}
	return authenticator.completeReauthentication(
		ctx, scope, request.Client, credential, proof,
	)
}

package authentication

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (authenticator Authenticator) RevokeSession(
	ctx context.Context,
	scope platformtx.CredentialTenantScope,
	request LogoutRequest,
) error {
	if request.Presented != nil {
		defer request.Presented.Destroy()
	}
	if !authenticator.valid || request.validate() != nil ||
		!validCredentialScope(
			ctx, scope, platformtx.CredentialRefresh,
			platformtx.CredentialRefreshLocator,
		) {
		return ErrAuthenticationContract
	}
	credential, err := authenticator.verifiedRefresh(ctx, scope, request.Presented)
	if err != nil {
		return err
	}
	defer clear(credential.SecretHMAC[:])
	result, err := authenticator.port.RevokeSession(
		ctx, scope, credential.logoutMutation(),
	)
	if err != nil || result.validate() != nil {
		return ErrAuthenticationStorage
	}
	return nil
}

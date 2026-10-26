package authentication

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (authenticator Authenticator) verifiedRefresh(
	ctx context.Context,
	scope platformtx.CredentialTenantScope,
	presented PresentedRefresh,
) (RefreshCredential, error) {
	credential, err := authenticator.port.RefreshCredential(ctx, scope)
	if err != nil || credential.validate(scope) != nil {
		return RefreshCredential{}, ErrInvalidCredentials
	}
	selector, err := presented.Selector()
	if err != nil || selector != scope.SourceSelector() {
		clear(credential.SecretHMAC[:])
		return RefreshCredential{}, ErrInvalidCredentials
	}
	key, err := authenticator.keys.HMACKey(credential.HashKeyID)
	if err != nil {
		clear(credential.SecretHMAC[:])
		return RefreshCredential{}, ErrAuthenticationCrypto
	}
	matched, verifyErr := presented.VerifyHMAC(key[:], credential.SecretHMAC)
	clear(key[:])
	if verifyErr != nil {
		clear(credential.SecretHMAC[:])
		return RefreshCredential{}, ErrAuthenticationCrypto
	}
	if !matched {
		clear(credential.SecretHMAC[:])
		return RefreshCredential{}, ErrInvalidCredentials
	}
	return credential, nil
}

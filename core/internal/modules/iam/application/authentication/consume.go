package authentication

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (authenticator Authenticator) ConsumeRecovery(
	ctx context.Context,
	scope platformtx.CredentialTenantScope,
	request RecoveryRequest,
) (int64, error) {
	if request.Code != nil {
		defer request.Code.Destroy()
	}
	if !authenticator.valid || request.validate() != nil ||
		!validCredentialScope(
			ctx, scope, platformtx.CredentialPasswordReset,
			platformtx.CredentialTenantRoute,
		) {
		return 0, ErrAuthenticationContract
	}
	keyID, err := request.Code.KeyID()
	if err != nil || !validHMACKeyID(keyID) {
		return 0, ErrInvalidCredentials
	}
	key, err := authenticator.security.RecoveryKeys.HMACKey(keyID)
	if err != nil {
		return 0, ErrAuthenticationCrypto
	}
	codeHMAC, err := request.Code.HMAC(key[:])
	clear(key[:])
	if err != nil {
		return 0, ErrAuthenticationCrypto
	}
	defer clear(codeHMAC[:])
	if codeHMAC != scope.LocatorBindingSHA256() {
		return 0, ErrAuthenticationContract
	}
	lookup := RecoveryLookup{HMACKeyID: keyID, CodeHMAC: codeHMAC}
	credential, err := authenticator.port.RecoveryCredential(ctx, scope, lookup)
	if err != nil || credential.validate(lookup) != nil {
		return 0, ErrInvalidCredentials
	}
	defer clear(credential.CodeHMAC[:])
	consumed, err := authenticator.port.ConsumeRecovery(
		ctx, scope, credential.consumption(),
	)
	if err != nil || consumed.validate(credential.TokenVersion) != nil {
		return 0, ErrAuthenticationStorage
	}
	return consumed.TokenVersion, nil
}

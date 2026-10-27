package authentication

import (
	"context"
	"time"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

// ValidateRefresh authenticates an active browser refresh credential without
// rotating it. It is intentionally read-only and is used only for CSRF proof.
func (authenticator Authenticator) ValidateRefresh(
	ctx context.Context,
	scope platformtx.CredentialTenantScope,
	presented PresentedRefresh,
) error {
	if !nilInterface(presented) {
		defer presented.Destroy()
	}
	if !authenticator.valid || nilInterface(presented) ||
		!validCredentialScope(
			ctx, scope, platformtx.CredentialRefresh,
			platformtx.CredentialRefreshLocator,
		) {
		return ErrAuthenticationContract
	}
	credential, err := authenticator.verifiedRefresh(ctx, scope, presented)
	if err != nil {
		return err
	}
	defer clear(credential.SecretHMAC[:])
	if !credential.activeAt(authenticator.clock().UTC()) {
		return ErrInvalidCredentials
	}
	return nil
}

func (credential RefreshCredential) activeAt(now time.Time) bool {
	if credential.TokenState != "active" || credential.FamilyState != "active" ||
		credential.SessionState != "active" || credential.UserStatus != "active" ||
		credential.AccessModeState != "active" || !now.Before(credential.IdleExpiresAt) ||
		!now.Before(credential.AbsoluteExpiresAt) {
		return false
	}
	return credential.AccessMode != AccessModePortal ||
		credential.PortalBindingState == "active"
}

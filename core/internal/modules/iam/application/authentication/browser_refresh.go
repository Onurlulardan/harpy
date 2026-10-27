package authentication

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

// rotateBrowser derives the public browser client from the authoritative
// access-mode row. Callers cannot select or upgrade the client identity.
func (authenticator Authenticator) rotateBrowser(
	ctx context.Context,
	scope platformtx.CredentialTenantScope,
	presented PresentedRefresh,
) (Rotation, error) {
	if !nilInterface(presented) {
		defer presented.Destroy()
	}
	if !authenticator.valid || nilInterface(presented) ||
		!validCredentialScope(
			ctx, scope, platformtx.CredentialRefresh,
			platformtx.CredentialRefreshLocator,
		) {
		return Rotation{}, ErrAuthenticationContract
	}
	credential, err := authenticator.verifiedRefresh(ctx, scope, presented)
	if err != nil {
		return Rotation{}, err
	}
	defer clear(credential.SecretHMAC[:])
	client := browserClientForMode(credential.AccessMode)
	if client == "" {
		return Rotation{}, ErrInvalidCredentials
	}
	accessContext, err := credential.accessContext()
	if err != nil {
		return Rotation{}, err
	}
	return authenticator.rotateVerified(ctx, scope, RotateRequest{
		Presented: presented, Client: client,
	}, credential, accessContext)
}

func browserClientForMode(mode AccessMode) ClientID {
	switch mode {
	case AccessModeOperator:
		return ClientOperatorConsole
	case AccessModePortal:
		return ClientPortal
	default:
		return ""
	}
}

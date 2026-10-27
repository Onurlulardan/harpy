package authentication

import (
	"context"
	"time"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (authenticator Authenticator) RotateBrowser(
	ctx context.Context,
	scope platformtx.CredentialTenantScope,
	presented PresentedRefresh,
) (Rotation, error) {
	return authenticator.rotateBrowser(ctx, scope, presented)
}

func (authenticator Authenticator) rotateWithClient(
	ctx context.Context,
	scope platformtx.CredentialTenantScope,
	request RotateRequest,
) (Rotation, error) {
	if request.Presented != nil {
		defer request.Presented.Destroy()
	}
	if !authenticator.valid || request.validate() != nil ||
		!validCredentialScope(
			ctx, scope, platformtx.CredentialRefresh,
			platformtx.CredentialRefreshLocator,
		) {
		return Rotation{}, ErrAuthenticationContract
	}
	credential, err := authenticator.verifiedRefresh(ctx, scope, request.Presented)
	if err != nil || !clientMatchesMode(request.Client, credential.AccessMode) {
		return Rotation{}, ErrInvalidCredentials
	}
	defer clear(credential.SecretHMAC[:])
	accessContext, err := credential.accessContext()
	if err != nil {
		return Rotation{}, err
	}
	return authenticator.rotateVerified(ctx, scope, request, credential, accessContext)
}

func (authenticator Authenticator) rotateVerified(
	ctx context.Context,
	scope platformtx.CredentialTenantScope,
	request RotateRequest,
	credential RefreshCredential,
	accessContext refreshAccessContext,
) (Rotation, error) {
	successor, err := authenticator.refresh.Generate()
	if err != nil || successor == nil {
		return Rotation{}, ErrAuthenticationCrypto
	}
	destroySuccessor := true
	defer func() {
		if destroySuccessor {
			successor.Destroy()
		}
	}()
	selector, err := successor.Selector()
	if err != nil {
		return Rotation{}, ErrAuthenticationCrypto
	}
	keyID, key, err := authenticator.keys.Active()
	if err != nil {
		return Rotation{}, ErrAuthenticationCrypto
	}
	secretHMAC, hmacErr := successor.SecretHMAC(key[:])
	clear(key[:])
	if hmacErr != nil {
		return Rotation{}, ErrAuthenticationCrypto
	}
	defer clear(secretHMAC[:])
	identifiers, err := authenticator.nextRotationIdentifiers()
	if err != nil {
		return Rotation{}, err
	}
	rotated, err := authenticator.port.Rotate(ctx, scope, credential.rotation(
		identifiers.token, selector, keyID, secretHMAC,
	))
	if err != nil || rotated.validate() != nil {
		return Rotation{}, ErrAuthenticationStorage
	}
	if rotated.Outcome != RefreshRotatedOutcome {
		return newTerminalRotation(rotated.Outcome)
	}
	issuedAt := rotated.IssuedAt.UTC().Truncate(time.Second)
	input := credential.claimsInput(
		scope.TenantID(), request.Client, identifiers.jti, issuedAt, accessContext,
	)
	access, err := authenticator.signAccess(input)
	if err != nil {
		return Rotation{}, err
	}
	result, err := newRotatedRotation(
		access, successor, credential.SessionID, credential.AbsoluteExpiresAt,
	)
	if err != nil {
		return Rotation{}, err
	}
	destroySuccessor = false
	return result, nil
}

func (credential RefreshCredential) claimsInput(
	tenantID uuid.UUID,
	client ClientID,
	jti uuid.UUID,
	issuedAt time.Time,
	context refreshAccessContext,
) AccessClaims {
	input := AccessClaims{
		Subject: credential.UserID, Session: credential.SessionID,
		Tenant: tenantID, Client: client, IssuedAt: issuedAt, JTI: jti,
		TokenVersion:               credential.TokenVersion,
		AuthorizationVersion:       credential.AuthorizationVersion,
		PlatformAuthorizationClock: credential.PlatformAuthorizationClock,
		SecurityEpoch:              credential.SecurityEpoch,
		AccessMode:                 credential.AccessMode,
		AuthenticationTime:         credential.AuthTime.UTC().Truncate(time.Second),
		Assurance:                  context.assurance, Methods: context.methods,
		StepUpPurposeHash: context.stepUp,
	}
	if credential.PortalBindingID != uuid.Nil {
		binding := credential.PortalBindingID
		input.PortalBinding = &binding
	}
	return input
}

package authentication

import (
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (request BeginReauthenticationRequest) matchesScope(
	scope platformtx.CredentialTenantScope,
) bool {
	binding, err := ReauthenticationBindingSHA256(
		scope.TenantID(), request.UserID, request.SessionID,
		request.TokenVersion, request.AuthorizationVersion,
		request.AccessMode, request.PortalBindingID,
	)
	return err == nil && binding == scope.LocatorBindingSHA256()
}

func (request BeginReauthenticationRequest) contextSHA256(
	tenantID uuid.UUID,
) ([32]byte, error) {
	return ReauthenticationContextSHA256(
		tenantID, request.UserID, request.SessionID, request.PurposeSHA256,
		request.ActionPolicyVersion, request.Method, request.OriginSHA256,
	)
}

func (request BeginReauthenticationRequest) creation(
	id uuid.UUID,
	keyID string,
	challengeHMAC, contextSHA [32]byte,
) ChallengeCreation {
	return ChallengeCreation{
		ID: id, UserID: request.UserID, SessionID: request.SessionID,
		ExpectedTokenVersion:         request.TokenVersion,
		ExpectedAuthorizationVersion: request.AuthorizationVersion,
		AccessMode:                   request.AccessMode, PortalBindingID: request.PortalBindingID,
		Method: request.Method, AuthenticatorID: request.AuthenticatorID,
		HMACKeyID: keyID, ChallengeHMAC: challengeHMAC,
		ContextSHA256: contextSHA, PurposeSHA256: request.PurposeSHA256,
		ActionPolicyVersion:      request.ActionPolicyVersion,
		OriginSHA256:             request.OriginSHA256,
		RequirePhishingResistant: request.RequirePhishingResistant,
	}
}

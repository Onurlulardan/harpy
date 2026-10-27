package authentication

import (
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type ReauthenticationLookup struct {
	ChallengeID, UserID, SessionID                     uuid.UUID
	ExpectedTokenVersion, ExpectedAuthorizationVersion int64
	AccessMode                                         AccessMode
	PortalBindingID                                    uuid.UUID
	Method                                             ReauthenticationMethod
	PurposeSHA256                                      [32]byte
	ActionPolicyVersion                                int64
	OriginSHA256                                       [32]byte
	ContextSHA256                                      [32]byte
}

func (request ReauthenticateRequest) lookup(
	scope platformtx.CredentialTenantScope,
) (ReauthenticationLookup, error) {
	binding, err := ReauthenticationBindingSHA256(
		scope.TenantID(), request.UserID, request.SessionID,
		request.TokenVersion, request.AuthorizationVersion,
		request.AccessMode, request.PortalBindingID,
	)
	if err != nil || binding != scope.LocatorBindingSHA256() {
		return ReauthenticationLookup{}, ErrAuthenticationContract
	}
	contextSHA, err := ReauthenticationContextSHA256(
		scope.TenantID(), request.UserID, request.SessionID,
		request.PurposeSHA256, request.ActionPolicyVersion,
		request.method(), request.OriginSHA256,
	)
	if err != nil {
		return ReauthenticationLookup{}, err
	}
	return ReauthenticationLookup{
		ChallengeID: request.ChallengeID, UserID: request.UserID,
		SessionID: request.SessionID, ExpectedTokenVersion: request.TokenVersion,
		ExpectedAuthorizationVersion: request.AuthorizationVersion,
		AccessMode:                   request.AccessMode, PortalBindingID: request.PortalBindingID,
		Method: request.method(), PurposeSHA256: request.PurposeSHA256,
		ActionPolicyVersion: request.ActionPolicyVersion,
		OriginSHA256:        request.OriginSHA256, ContextSHA256: contextSHA,
	}, nil
}

package authentication

import (
	"crypto/sha256"
	"strings"

	"github.com/google/uuid"
)

type BeginReauthenticationRequest struct {
	UserID, SessionID                  uuid.UUID
	TokenVersion, AuthorizationVersion int64
	AccessMode                         AccessMode
	PortalBindingID                    uuid.UUID
	Client                             ClientID
	Method                             ReauthenticationMethod
	AuthenticatorID                    uuid.UUID
	PurposeSHA256                      [32]byte
	ActionPolicyVersion                int64
	Origin                             string
	OriginSHA256                       [32]byte
	RequirePhishingResistant           bool
}

func (request BeginReauthenticationRequest) validate() error {
	originDigest := sha256.Sum256([]byte(request.Origin))
	authenticatorShape := request.Method == ReauthenticationPassword &&
		request.AuthenticatorID == uuid.Nil ||
		(request.Method == ReauthenticationWebAuthn ||
			request.Method == ReauthenticationTOTP) && request.AuthenticatorID != uuid.Nil
	if request.UserID == uuid.Nil || request.SessionID == uuid.Nil ||
		request.TokenVersion <= 0 || request.AuthorizationVersion <= 0 ||
		!clientMatchesMode(request.Client, request.AccessMode) ||
		(request.AccessMode == AccessModeOperator && request.PortalBindingID != uuid.Nil) ||
		(request.AccessMode == AccessModePortal && request.PortalBindingID == uuid.Nil) ||
		!request.Method.valid() || !authenticatorShape ||
		request.PurposeSHA256 == [32]byte{} || request.ActionPolicyVersion <= 0 ||
		request.Origin == "" || len(request.Origin) > 2048 ||
		strings.TrimSpace(request.Origin) != request.Origin ||
		request.OriginSHA256 == [32]byte{} || request.OriginSHA256 != originDigest ||
		(request.RequirePhishingResistant && request.Method != ReauthenticationWebAuthn) {
		return ErrAuthenticationContract
	}
	return nil
}

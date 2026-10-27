package authentication

import (
	"crypto/sha256"
	"strings"

	"github.com/google/uuid"
)

type ReauthenticateRequest struct {
	ChallengeID                        uuid.UUID
	Challenge                          PresentedChallenge
	Password                           PasswordSecret
	WebAuthn                           WebAuthnAssertion
	TOTP                               TOTPCode
	UserID, SessionID                  uuid.UUID
	TokenVersion, AuthorizationVersion int64
	AccessMode                         AccessMode
	PortalBindingID                    uuid.UUID
	Client                             ClientID
	PurposeSHA256                      [32]byte
	ActionPolicyVersion                int64
	Origin                             string
	OriginSHA256                       [32]byte
}

func (request ReauthenticateRequest) validate() error {
	originDigest := sha256.Sum256([]byte(request.Origin))
	proofs := 0
	for _, present := range []bool{
		!nilInterface(request.Password), !nilInterface(request.WebAuthn),
		!nilInterface(request.TOTP),
	} {
		if present {
			proofs++
		}
	}
	if request.ChallengeID == uuid.Nil || nilInterface(request.Challenge) ||
		proofs != 1 || request.UserID == uuid.Nil || request.SessionID == uuid.Nil ||
		request.TokenVersion <= 0 || request.AuthorizationVersion <= 0 ||
		!clientMatchesMode(request.Client, request.AccessMode) ||
		(request.AccessMode == AccessModeOperator && request.PortalBindingID != uuid.Nil) ||
		(request.AccessMode == AccessModePortal && request.PortalBindingID == uuid.Nil) ||
		request.PurposeSHA256 == [32]byte{} || request.ActionPolicyVersion <= 0 ||
		request.Origin == "" || len(request.Origin) > 2048 ||
		strings.TrimSpace(request.Origin) != request.Origin ||
		request.OriginSHA256 == [32]byte{} || request.OriginSHA256 != originDigest {
		return ErrAuthenticationContract
	}
	return nil
}

func (request ReauthenticateRequest) method() ReauthenticationMethod {
	switch {
	case !nilInterface(request.Password):
		return ReauthenticationPassword
	case !nilInterface(request.WebAuthn):
		return ReauthenticationWebAuthn
	case !nilInterface(request.TOTP):
		return ReauthenticationTOTP
	default:
		return ""
	}
}

func (request *ReauthenticateRequest) destroy() {
	if request == nil {
		return
	}
	if !nilInterface(request.Challenge) {
		request.Challenge.Destroy()
	}
	if !nilInterface(request.Password) {
		request.Password.Destroy()
	}
	if !nilInterface(request.WebAuthn) {
		request.WebAuthn.Destroy()
	}
	if !nilInterface(request.TOTP) {
		request.TOTP.Destroy()
	}
}

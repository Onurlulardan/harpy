package iam

import (
	"encoding/base64"
	"fmt"
	"time"
)

type authHTTPReauthenticationWebAuthn struct {
	CredentialID             string `json:"credential_id"`
	RPID                     string `json:"rp_id"`
	UserVerificationRequired bool   `json:"user_verification_required"`
}

type authHTTPReauthenticationResponse struct {
	ChallengeID         string                            `json:"challenge_id"`
	Method              AuthHTTPReauthenticationMethod    `json:"method"`
	Challenge           string                            `json:"challenge"`
	ExpiresAt           time.Time                         `json:"expires_at"`
	ActionPolicyVersion int64                             `json:"action_policy_version"`
	WebAuthn            *authHTTPReauthenticationWebAuthn `json:"webauthn"`
}

func reauthenticationResponse(
	challenge AuthHTTPReauthenticationChallenge,
) (authHTTPReauthenticationResponse, error) {
	if challenge == nil || challenge.ID() == [16]byte{} ||
		challenge.ExpiresAt().IsZero() {
		return authHTTPReauthenticationResponse{}, fmt.Errorf("challenge is invalid")
	}
	response := authHTTPReauthenticationResponse{ChallengeID: challenge.ID().String(),
		Method: challenge.Method(), ExpiresAt: challenge.ExpiresAt().UTC(),
		ActionPolicyVersion: AuthHTTPReauthenticationPolicy}
	err := challenge.RevealOnce(func(wire []byte) error {
		if !canonicalBase64URL(string(wire), 32) {
			return fmt.Errorf("challenge material is invalid")
		}
		response.Challenge = string(wire)
		return nil
	})
	if err != nil || response.Challenge == "" {
		return authHTTPReauthenticationResponse{}, fmt.Errorf("challenge reveal failed")
	}
	if response.Method == AuthHTTPReauthenticationWebAuthn {
		credentialID := challenge.CredentialID()
		defer clear(credentialID)
		if len(credentialID) == 0 || challenge.RPID() == "" {
			return authHTTPReauthenticationResponse{}, fmt.Errorf("WebAuthn options are invalid")
		}
		response.WebAuthn = &authHTTPReauthenticationWebAuthn{
			CredentialID:             base64.RawURLEncoding.EncodeToString(credentialID),
			RPID:                     challenge.RPID(),
			UserVerificationRequired: challenge.UserVerificationRequired(),
		}
	} else if len(challenge.CredentialID()) != 0 || challenge.RPID() != "" {
		return authHTTPReauthenticationResponse{}, fmt.Errorf("challenge method differs")
	}
	return response, nil
}

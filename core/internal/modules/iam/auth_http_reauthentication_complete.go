package iam

import (
	"context"
	"crypto/sha256"

	cryptography "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/cryptography"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authentication"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/password"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *authHTTPApplication) CompleteReauthentication(
	ctx context.Context,
	claims accesstoken.Claims,
	input AuthHTTPReauthenticationCompleteInput,
) (AuthHTTPGrant, error) {
	defer clearReauthenticationCompleteInput(&input)
	if application == nil || ctx == nil || ctx.Err() != nil ||
		!validReauthenticationCompleteInput(input) {
		return nil, ErrAuthHTTPInvalidSession
	}
	session, err := reauthenticationSessionFromClaims(claims)
	if err != nil {
		return nil, ErrAuthHTTPInvalidSession
	}
	lookup, err := session.lookup()
	if err != nil {
		return nil, ErrAuthHTTPUnavailable
	}
	request, err := buildReauthenticationRequest(input, session)
	if err != nil {
		return nil, ErrAuthHTTPInvalidCredentials
	}
	consumed := false
	defer func() {
		if !consumed {
			destroyReauthenticationRequest(&request)
		}
	}()
	var grant authentication.Grant
	err = application.coordinator.InCredentialTenant(ctx,
		platformtx.CredentialScopeRequest{Lookup: lookup, Resolver: application.resolver},
		func(scoped context.Context, scope platformtx.CredentialTenantScope) error {
			consumed = true
			var completeErr error
			grant, completeErr = application.authenticator.CompleteReauthentication(
				scoped, scope, request,
			)
			return completeErr
		})
	if err != nil {
		grant.Destroy()
		return nil, mapHTTPSessionError(err)
	}
	return &grant, nil
}

func destroyReauthenticationRequest(request *authentication.ReauthenticateRequest) {
	if request == nil {
		return
	}
	if request.Challenge != nil {
		request.Challenge.Destroy()
	}
	if request.Password != nil {
		request.Password.Destroy()
	}
	if request.WebAuthn != nil {
		request.WebAuthn.Destroy()
	}
	if request.TOTP != nil {
		request.TOTP.Destroy()
	}
}

func buildReauthenticationRequest(
	input AuthHTTPReauthenticationCompleteInput,
	session reauthenticationSession,
) (authentication.ReauthenticateRequest, error) {
	challenge, err := cryptography.NewPresentedChallenge(input.Challenge)
	if err != nil {
		return authentication.ReauthenticateRequest{}, err
	}
	request := authentication.ReauthenticateRequest{ChallengeID: input.ChallengeID,
		Challenge: challenge, UserID: session.userID, SessionID: session.sessionID,
		TokenVersion: session.tokenVersion, AuthorizationVersion: session.authorizationVersion,
		AccessMode: session.accessMode, PortalBindingID: session.portalBindingID,
		Client: session.client, PurposeSHA256: input.PurposeSHA256,
		ActionPolicyVersion: AuthHTTPReauthenticationPolicy, Origin: input.Origin,
		OriginSHA256: sha256.Sum256([]byte(input.Origin))}
	if err := attachReauthenticationProof(&request, input); err != nil {
		challenge.Destroy()
		return authentication.ReauthenticateRequest{}, err
	}
	return request, nil
}

func attachReauthenticationProof(
	request *authentication.ReauthenticateRequest,
	input AuthHTTPReauthenticationCompleteInput,
) error {
	var err error
	switch input.Method {
	case AuthHTTPReauthenticationPassword:
		request.Password, err = password.NewVerificationSecret(input.Password)
	case AuthHTTPReauthenticationWebAuthn:
		request.WebAuthn, err = cryptography.NewWebAuthnAssertion(input.WebAuthnAssertion)
	case AuthHTTPReauthenticationTOTP:
		request.TOTP, err = cryptography.NewTOTPCode(input.TOTP)
	default:
		return authentication.ErrAuthenticationContract
	}
	return err
}

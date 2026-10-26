package iam

import (
	"context"
	"crypto/sha256"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authentication"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *authHTTPApplication) StartReauthentication(
	ctx context.Context,
	claims accesstoken.Claims,
	input AuthHTTPReauthenticationStartInput,
) (AuthHTTPReauthenticationChallenge, error) {
	if application == nil || ctx == nil || ctx.Err() != nil ||
		!validReauthenticationStartInput(input) {
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
	request := authentication.BeginReauthenticationRequest{
		UserID: session.userID, SessionID: session.sessionID,
		TokenVersion: session.tokenVersion, AuthorizationVersion: session.authorizationVersion,
		AccessMode: session.accessMode, PortalBindingID: session.portalBindingID,
		Client: session.client, Method: authentication.ReauthenticationMethod(input.Method),
		AuthenticatorID: input.AuthenticatorID, PurposeSHA256: input.PurposeSHA256,
		ActionPolicyVersion: AuthHTTPReauthenticationPolicy, Origin: input.Origin,
		OriginSHA256: sha256.Sum256([]byte(input.Origin)),
	}
	var challenge authentication.ReauthenticationChallenge
	err = application.coordinator.InCredentialTenant(ctx,
		platformtx.CredentialScopeRequest{Lookup: lookup, Resolver: application.resolver},
		func(scoped context.Context, scope platformtx.CredentialTenantScope) error {
			var startErr error
			challenge, startErr = application.authenticator.StartReauthentication(
				scoped, scope, request,
			)
			return startErr
		})
	if err != nil {
		challenge.Destroy()
		return nil, mapHTTPSessionError(err)
	}
	return &authHTTPReauthenticationChallenge{inner: &challenge}, nil
}

func validReauthenticationStartInput(input AuthHTTPReauthenticationStartInput) bool {
	method := authentication.ReauthenticationMethod(input.Method)
	authenticatorShape := method == authentication.ReauthenticationPassword &&
		input.AuthenticatorID == [16]byte{} ||
		(method == authentication.ReauthenticationWebAuthn ||
			method == authentication.ReauthenticationTOTP) && input.AuthenticatorID != [16]byte{}
	return authenticatorShape && input.PurposeSHA256 != [32]byte{} &&
		input.Origin != "" && len(input.Origin) <= 512
}

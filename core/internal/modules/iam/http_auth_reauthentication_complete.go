package iam

import (
	"crypto/subtle"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpwire"
)

func (handler *AuthHTTP) completeReauthentication(w http.ResponseWriter, r *http.Request) {
	origin, valid := exactOrigin(r)
	if !valid || !handler.security.browserOriginAllowed(origin) {
		handler.writeAuthFailure(w, r, authFailureCSRF)
		return
	}
	input, err := decodeReauthenticationComplete(w, r, origin)
	if err != nil {
		handler.writeInputFailure(w, r, err)
		return
	}
	defer clearReauthenticationCompleteInput(&input)
	claims, valid := handler.reauthenticationClaims(w, r)
	if !valid {
		return
	}
	grant, err := handler.application.CompleteReauthentication(r.Context(), claims, input)
	if err != nil {
		handler.writeApplicationFailure(w, r, err)
		return
	}
	if nilAuthHTTPGrant(grant) {
		handler.report(errors.New("IAM reauthentication returned no grant"))
		handler.writeAuthFailure(w, r, authFailureUnavailable)
		return
	}
	defer grant.Destroy()
	now, valid := handler.authHTTPNow()
	if !valid {
		handler.writeAuthFailure(w, r, authFailureUnavailable)
		return
	}
	access, err := grant.AccessToken()
	if err != nil || !handler.validStepUpGrant(r, access, claims, input, now) {
		handler.report(errors.New("IAM reauthentication grant differs from request"))
		handler.writeAuthFailure(w, r, authFailureUnavailable)
		return
	}
	response, cookies, err := handler.prepareGrant(r.Context(), grant, origin, now)
	if err != nil {
		handler.report(errors.New("IAM reauthentication grant could not be presented"))
		handler.writeAuthFailure(w, r, authFailureUnavailable)
		return
	}
	appendSetCookies(w.Header(), cookies)
	if err := httpwire.WriteJSON(w, http.StatusOK,
		authHTTPCorrelationID(r), response, false); err != nil {
		handler.report(fmt.Errorf("write IAM reauthentication grant: %w", err))
	}
}

func (handler *AuthHTTP) validStepUpGrant(
	r *http.Request, token string, baseline accesstoken.Claims,
	input AuthHTTPReauthenticationCompleteInput, now time.Time,
) bool {
	claims, err := handler.security.accessVerifier.Verify(r.Context(), token, now)
	purpose, present := claims.StepUpPurposeHash()
	portal, hasPortal := claims.PortalBinding()
	baselinePortal, baselineHasPortal := baseline.PortalBinding()
	methods := claims.Methods()
	return err == nil && claims.Assurance() == accesstoken.AssuranceStepUp && present &&
		subtle.ConstantTimeCompare(purpose[:], input.PurposeSHA256[:]) == 1 &&
		len(methods) == 1 && methods[0] == reauthenticationAuthMethod(input.Method) &&
		claims.Subject() == baseline.Subject() && claims.Tenant() == baseline.Tenant() &&
		claims.Session() == baseline.Session() && claims.Client() == baseline.Client() &&
		claims.AccessMode() == baseline.AccessMode() && claims.TokenVersion() == baseline.TokenVersion() &&
		claims.AuthorizationVersion() == baseline.AuthorizationVersion() &&
		claims.PlatformAuthorizationClock() == baseline.PlatformAuthorizationClock() &&
		claims.SecurityEpoch() == baseline.SecurityEpoch() &&
		hasPortal == baselineHasPortal && portal == baselinePortal
}

func reauthenticationAuthMethod(method AuthHTTPReauthenticationMethod) accesstoken.AuthMethod {
	switch method {
	case AuthHTTPReauthenticationPassword:
		return accesstoken.AuthMethodPassword
	case AuthHTTPReauthenticationWebAuthn:
		return accesstoken.AuthMethodWebAuthn
	case AuthHTTPReauthenticationTOTP:
		return accesstoken.AuthMethodTOTP
	default:
		return ""
	}
}

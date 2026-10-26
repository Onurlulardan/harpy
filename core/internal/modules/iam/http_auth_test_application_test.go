package iam

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authentication"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/refreshtoken"
)

type authHTTPTestApplication struct {
	initialStartInput              AuthHTTPLoginInput
	initialStartChallenge          AuthHTTPInitialChallenge
	initialStartError              error
	initialCompleteInput           AuthHTTPInitialPasswordInput
	initialCompleteGrant           AuthHTTPGrant
	initialCompleteError           error
	loginInput                     AuthHTTPLoginInput
	loginGrant                     AuthHTTPGrant
	loginError                     error
	refreshGrant                   AuthHTTPGrant
	refreshError                   error
	logoutError                    error
	validationError                error
	currentSession                 AuthHTTPSession
	currentError                   error
	reauthenticationStartInput     AuthHTTPReauthenticationStartInput
	reauthenticationStartChallenge AuthHTTPReauthenticationChallenge
	reauthenticationStartError     error
	reauthenticationCompleteInput  AuthHTTPReauthenticationCompleteInput
	reauthenticationCompleteGrant  AuthHTTPGrant
	reauthenticationCompleteError  error
	loginCalls                     int
	refreshCalls                   int
	logoutCalls                    int
	validationCalls                int
	currentCalls                   int
	reauthenticationStartCalls     int
	reauthenticationCompleteCalls  int
	initialStartCalls              int
	initialCompleteCalls           int
}

func (application *authHTTPTestApplication) StartReauthentication(
	_ context.Context,
	_ accesstoken.Claims,
	input AuthHTTPReauthenticationStartInput,
) (AuthHTTPReauthenticationChallenge, error) {
	application.reauthenticationStartCalls++
	application.reauthenticationStartInput = input
	return application.reauthenticationStartChallenge,
		application.reauthenticationStartError
}

func (application *authHTTPTestApplication) CompleteReauthentication(
	_ context.Context,
	_ accesstoken.Claims,
	input AuthHTTPReauthenticationCompleteInput,
) (AuthHTTPGrant, error) {
	application.reauthenticationCompleteCalls++
	application.reauthenticationCompleteInput = input
	return application.reauthenticationCompleteGrant,
		application.reauthenticationCompleteError
}

func (application *authHTTPTestApplication) StartInitialPassword(
	_ context.Context,
	input AuthHTTPLoginInput,
) (AuthHTTPInitialChallenge, error) {
	application.initialStartCalls++
	application.initialStartInput = input
	return application.initialStartChallenge, application.initialStartError
}

func (application *authHTTPTestApplication) CompleteInitialPassword(
	_ context.Context,
	input AuthHTTPInitialPasswordInput,
	_ *authentication.PresentedInitialPassword,
) (AuthHTTPGrant, error) {
	application.initialCompleteCalls++
	application.initialCompleteInput = input
	return application.initialCompleteGrant, application.initialCompleteError
}

func (application *authHTTPTestApplication) Login(
	_ context.Context,
	input AuthHTTPLoginInput,
) (AuthHTTPGrant, error) {
	application.loginCalls++
	application.loginInput = input
	return application.loginGrant, application.loginError
}

func (application *authHTTPTestApplication) Refresh(
	_ context.Context,
	_ *refreshtoken.Presented,
) (AuthHTTPGrant, error) {
	application.refreshCalls++
	return application.refreshGrant, application.refreshError
}

func (application *authHTTPTestApplication) Logout(
	_ context.Context,
	_ *refreshtoken.Presented,
) error {
	application.logoutCalls++
	return application.logoutError
}

func (application *authHTTPTestApplication) ValidateRefresh(
	_ context.Context,
	_ *refreshtoken.Presented,
) error {
	application.validationCalls++
	return application.validationError
}

func (application *authHTTPTestApplication) CurrentSession(
	_ context.Context,
	_ accesstoken.Claims,
) (AuthHTTPSession, error) {
	application.currentCalls++
	return application.currentSession, application.currentError
}

type authHTTPTestGrant struct {
	access    string
	refresh   *refreshtoken.Issued
	absolute  time.Time
	destroyed bool
}

func (grant *authHTTPTestGrant) AccessToken() (string, error) {
	return grant.access, nil
}

func (grant *authHTTPTestGrant) AbsoluteExpiresAt() time.Time {
	return grant.absolute
}

func (grant *authHTTPTestGrant) RevealRefreshOnce(consume func([]byte) error) error {
	return grant.refresh.RevealOnce(consume)
}

func (grant *authHTTPTestGrant) Destroy() {
	if grant == nil {
		return
	}
	if grant.refresh != nil {
		grant.refresh.Destroy()
	}
	grant.access = ""
	grant.destroyed = true
}

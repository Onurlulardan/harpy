package iam

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycle"
	"github.com/google/uuid"
)

type lifecycleHTTPTestApplication struct {
	err                    error
	listAuthenticatorCalls int
	sessions               []lifecycle.Session
	listSessionCalls       int
	portalSetup            lifecycle.PortalSetup
	readPortalSetupCalls   int
	readPortalSetupID      uuid.UUID
}

func (app *lifecycleHTTPTestApplication) AddIdentifier(
	context.Context, accesstoken.Claims, LifecycleIdentifierAddInput,
) (LifecycleMutationView, error) {
	return LifecycleMutationView{}, app.err
}

func (app *lifecycleHTTPTestApplication) StartEmailVerification(
	context.Context, accesstoken.Claims, LifecycleResourceInput,
) error {
	return app.err
}

func (app *lifecycleHTTPTestApplication) VerifyIdentifier(
	context.Context, LifecycleCredentialInput,
) error {
	return app.err
}

func (app *lifecycleHTTPTestApplication) TombstoneIdentifier(
	context.Context, accesstoken.Claims, LifecycleRevisionInput,
) error {
	return app.err
}

func (app *lifecycleHTTPTestApplication) RetireIdentifier(
	context.Context, accesstoken.Claims, LifecycleRevisionInput,
) error {
	return app.err
}

func (app *lifecycleHTTPTestApplication) StartPasswordReset(
	context.Context, LifecyclePasswordResetStartInput,
) error {
	return app.err
}

func (app *lifecycleHTTPTestApplication) CompletePasswordReset(
	context.Context, LifecyclePasswordResetCompleteInput,
) error {
	return app.err
}

func (app *lifecycleHTTPTestApplication) CompleteInitialAdminSetup(
	context.Context, LifecyclePasswordResetCompleteInput,
) error {
	return app.err
}

func (app *lifecycleHTTPTestApplication) ChangePassword(
	context.Context, accesstoken.Claims, LifecyclePasswordChangeInput,
) error {
	return app.err
}

func (app *lifecycleHTTPTestApplication) StartPortalSetup(
	context.Context, accesstoken.Claims, LifecycleResourceInput,
) error {
	return app.err
}

func (app *lifecycleHTTPTestApplication) CompletePortalSetup(
	context.Context, LifecyclePasswordResetCompleteInput,
) error {
	return app.err
}

func (app *lifecycleHTTPTestApplication) ReadPortalSetup(
	_ context.Context, _ accesstoken.Claims, bindingID uuid.UUID,
) (lifecycle.PortalSetup, error) {
	app.readPortalSetupCalls++
	app.readPortalSetupID = bindingID
	return app.portalSetup, app.err
}

func (app *lifecycleHTTPTestApplication) DecidePortalSetup(
	context.Context, accesstoken.Claims, LifecyclePortalDecisionInput,
) error {
	return app.err
}

func (app *lifecycleHTTPTestApplication) StartWebAuthn(
	context.Context, accesstoken.Claims, LifecycleEnrollmentStartInput,
) (LifecycleChallengeView, error) {
	return LifecycleChallengeView{}, app.err
}

func (app *lifecycleHTTPTestApplication) CompleteWebAuthn(
	context.Context, accesstoken.Claims, LifecycleWebAuthnCompleteInput,
) error {
	return app.err
}

func (app *lifecycleHTTPTestApplication) StartTOTP(
	context.Context, accesstoken.Claims, LifecycleEnrollmentStartInput,
) (LifecycleTOTPStartView, error) {
	return LifecycleTOTPStartView{}, app.err
}

func (app *lifecycleHTTPTestApplication) CompleteTOTP(
	context.Context, accesstoken.Claims, LifecycleTOTPCompleteInput,
) error {
	return app.err
}

func (app *lifecycleHTTPTestApplication) IssueRecoveryCodes(
	context.Context, accesstoken.Claims, LifecycleRecoveryIssueInput,
) (LifecycleRecoveryIssueView, error) {
	return LifecycleRecoveryIssueView{}, app.err
}

func (app *lifecycleHTTPTestApplication) ListRecoveryCodes(
	context.Context, accesstoken.Claims,
) ([]lifecycle.RecoveryCode, error) {
	return nil, app.err
}

func (app *lifecycleHTTPTestApplication) ListAuthenticators(
	context.Context, accesstoken.Claims, LifecyclePageInput,
) ([]lifecycle.Authenticator, error) {
	app.listAuthenticatorCalls++
	return []lifecycle.Authenticator{}, app.err
}

func (app *lifecycleHTTPTestApplication) SuspendAuthenticator(
	context.Context, accesstoken.Claims, LifecycleRevisionInput,
) error {
	return app.err
}

func (app *lifecycleHTTPTestApplication) RetireAuthenticator(
	context.Context, accesstoken.Claims, LifecycleRevisionInput,
) error {
	return app.err
}

func (app *lifecycleHTTPTestApplication) ListSessions(
	context.Context, accesstoken.Claims, LifecyclePageInput,
) ([]lifecycle.Session, error) {
	app.listSessionCalls++
	return app.sessions, app.err
}

func (app *lifecycleHTTPTestApplication) RevokeSession(
	context.Context, accesstoken.Claims, LifecycleRevisionInput,
) error {
	return app.err
}

func (app *lifecycleHTTPTestApplication) RevokeAllSessions(
	context.Context, accesstoken.Claims, string,
) error {
	return app.err
}

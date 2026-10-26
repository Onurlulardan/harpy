package iam

import (
	"context"
	"crypto/subtle"
	"time"

	cryptography "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/cryptography"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authentication"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycle"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *lifecycleHTTPApplication) CompleteTOTP(
	ctx context.Context, claims accesstoken.Claims, input LifecycleTOTPCompleteInput,
) error {
	defer clear(input.Challenge)
	defer clear(input.Code)
	if input.AuthenticatorID == uuid.Nil || input.ChallengeID == uuid.Nil ||
		input.ExpectedAuthenticatorRevision < 1 || input.ExpectedChallengeRevision < 1 ||
		input.ChallengeKeyID == "" || !identityIdempotencyPattern.MatchString(input.IdempotencyKey) {
		return ErrLifecycleHTTPInvalid
	}
	challenge, _, err := application.presentedChallenge(claims, input.AuthenticatorID,
		input.ChallengeID, "totp_v1", input.ChallengeKeyID, input.Challenge)
	if err != nil {
		return err
	}
	var enrollment lifecycle.TOTPEnrollment
	err = application.inSnapshot(ctx, claims, "read",
		func(scoped context.Context, scope platformtx.TenantScope, permissionID uuid.UUID) error {
			var readErr error
			enrollment, readErr = application.runtime.service.TOTPEnrollmentLookup(scoped, scope,
				lifecycle.TOTPReadQuery{AuthenticatorID: input.AuthenticatorID,
					ChallengeID: input.ChallengeID, ActorID: claims.Subject(), PermissionID: permissionID,
					ExpectedTokenVersion:         claims.TokenVersion(),
					ExpectedAuthorizationVersion: claims.AuthorizationVersion()})
			return readErr
		})
	if err != nil {
		return err
	}
	if enrollment.ChallengeHMACKeyID != challenge.HMACKeyID ||
		subtle.ConstantTimeCompare(enrollment.ChallengeHMAC[:], challenge.HMAC[:]) != 1 ||
		subtle.ConstantTimeCompare(enrollment.ContextSHA256[:], challenge.ContextSHA256[:]) != 1 {
		return ErrLifecycleHTTPInvalidCredential
	}
	code, err := cryptography.NewTOTPCode(input.Code)
	if err != nil {
		return ErrLifecycleHTTPInvalidCredential
	}
	defer code.Destroy()
	step, err := application.totp.Verify(ctx, code, authentication.TOTPEnvelope{
		KMSKeyID: enrollment.KMSKeyID, Nonce: enrollment.Nonce,
		Ciphertext: enrollment.Ciphertext, AuthenticationTag: enrollment.AuthenticationTag,
		AADSHA256: enrollment.AADSHA256}, time.Now().UTC())
	if err != nil {
		return ErrLifecycleHTTPInvalidCredential
	}
	operationID := lifecycleOperationID(claims, "iam.totp.enrollment.complete", input.IdempotencyKey)
	requestSHA, _ := lifecycleRequestSHA(struct{ AuthenticatorID, ChallengeID uuid.UUID }{
		input.AuthenticatorID, input.ChallengeID})
	return application.inCommand(ctx, claims, "iam.totp.enrollment.complete", "update",
		func(scoped context.Context, scope platformtx.TenantScope, permissionID uuid.UUID) error {
			_, executeErr := application.runtime.service.ActivateTOTP(scoped, scope,
				lifecycle.TOTPCompleteCommand{
					Evidence:        lifecycleEvidence(claims, permissionID, operationID, requestSHA),
					AuthenticatorID: input.AuthenticatorID, Challenge: challenge,
					ExpectedAuthenticatorRevision: input.ExpectedAuthenticatorRevision,
					ExpectedChallengeRevision:     input.ExpectedChallengeRevision,
					AcceptedTimeStep:              step,
				})
			return executeErr
		})
}

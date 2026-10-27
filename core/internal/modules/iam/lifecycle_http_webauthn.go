package iam

import (
	"context"
	"crypto/sha256"

	cryptography "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/cryptography"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycle"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *lifecycleHTTPApplication) StartWebAuthn(
	ctx context.Context, claims accesstoken.Claims, input LifecycleEnrollmentStartInput,
) (LifecycleChallengeView, error) {
	if input.DisplayLabel == "" || !identityIdempotencyPattern.MatchString(input.IdempotencyKey) {
		return LifecycleChallengeView{}, ErrLifecycleHTTPInvalid
	}
	operationID := lifecycleOperationID(claims, "iam.webauthn.enrollment.start", input.IdempotencyKey)
	authenticatorID, challengeID := lifecycleChildID(operationID, "authenticator"), lifecycleChildID(operationID, "challenge")
	challenge, wire, err := application.issueChallenge(ctx, claims,
		authenticatorID, challengeID, "webauthn_v1")
	if err != nil {
		return LifecycleChallengeView{}, err
	}
	requestSHA, _ := lifecycleRequestSHA(struct{ Label string }{input.DisplayLabel})
	err = application.inCommand(ctx, claims, "iam.webauthn.enrollment.start", "update",
		func(scoped context.Context, scope platformtx.TenantScope, permissionID uuid.UUID) error {
			_, executeErr := application.runtime.service.EnableWebAuthn(scoped, scope,
				lifecycle.WebAuthnStartCommand{
					Evidence:        lifecycleEvidence(claims, permissionID, operationID, requestSHA),
					AuthenticatorID: authenticatorID, UserID: claims.Subject(), SessionID: claims.Session(),
					DisplayLabel: input.DisplayLabel, Challenge: challenge,
					OriginPolicyVersion: 1, RPIDSHA256: sha256.Sum256([]byte(application.rpID)),
					OriginSHA256: sha256.Sum256([]byte(application.origin)), UserVerificationRequired: true,
				})
			return executeErr
		})
	if err != nil {
		return LifecycleChallengeView{}, err
	}
	return LifecycleChallengeView{AuthenticatorID: authenticatorID,
		ChallengeID: challengeID, Challenge: wire, ChallengeKeyID: challenge.HMACKeyID,
		RPID: application.rpID, Origin: application.origin, UserVerificationRequired: true}, nil
}

func (application *lifecycleHTTPApplication) CompleteWebAuthn(
	ctx context.Context, claims accesstoken.Claims, input LifecycleWebAuthnCompleteInput,
) error {
	defer clear(input.Challenge)
	defer clear(input.Credential)
	if input.AuthenticatorID == uuid.Nil || input.ChallengeID == uuid.Nil ||
		input.ExpectedAuthenticatorRevision < 1 || input.ChallengeKeyID == "" ||
		!identityIdempotencyPattern.MatchString(input.IdempotencyKey) {
		return ErrLifecycleHTTPInvalid
	}
	challenge, wire, err := application.presentedChallenge(claims, input.AuthenticatorID,
		input.ChallengeID, "webauthn_v1", input.ChallengeKeyID, input.Challenge)
	if err != nil {
		return err
	}
	result, err := cryptography.VerifyWebAuthnRegistration(ctx, input.Credential,
		cryptography.WebAuthnRegistrationExpectation{Challenge: wire,
			RPID: application.rpID, Origin: application.origin, UserVerificationRequired: true})
	if err != nil {
		return ErrLifecycleHTTPInvalidCredential
	}
	operationID := lifecycleOperationID(claims, "iam.webauthn.enrollment.complete", input.IdempotencyKey)
	requestSHA, _ := lifecycleRequestSHA(struct{ AuthenticatorID, ChallengeID uuid.UUID }{
		input.AuthenticatorID, input.ChallengeID})
	return application.inCommand(ctx, claims, "iam.webauthn.enrollment.complete", "update",
		func(scoped context.Context, scope platformtx.TenantScope, permissionID uuid.UUID) error {
			_, executeErr := application.runtime.service.FinalizeWebAuthn(scoped, scope,
				lifecycle.WebAuthnCompleteCommand{
					Evidence:        lifecycleEvidence(claims, permissionID, operationID, requestSHA),
					AuthenticatorID: input.AuthenticatorID, Challenge: challenge,
					CredentialID: result.CredentialID, PublicKeyCOSE: result.PublicKeyCOSE,
					RPID: application.rpID, RPIDSHA256: sha256.Sum256([]byte(application.rpID)),
					OriginPolicyVersion: 1, AAGUID: result.AAGUID, SignCount: result.SignCount,
					BackupEligible: result.BackupEligible, BackupState: result.BackupState,
					Discoverable: result.Discoverable, UserVerificationRequired: true,
				})
			return executeErr
		})
}

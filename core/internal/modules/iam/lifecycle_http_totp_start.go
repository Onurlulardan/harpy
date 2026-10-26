package iam

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"io"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycle"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *lifecycleHTTPApplication) StartTOTP(
	ctx context.Context, claims accesstoken.Claims, input LifecycleEnrollmentStartInput,
) (LifecycleTOTPStartView, error) {
	if input.DisplayLabel == "" || !identityIdempotencyPattern.MatchString(input.IdempotencyKey) {
		return LifecycleTOTPStartView{}, ErrLifecycleHTTPInvalid
	}
	operationID := lifecycleOperationID(claims, "iam.totp.enrollment.start", input.IdempotencyKey)
	authenticatorID, challengeID := lifecycleChildID(operationID, "authenticator"), lifecycleChildID(operationID, "challenge")
	challenge, wire, err := application.issueChallenge(ctx, claims,
		authenticatorID, challengeID, "totp_v1")
	if err != nil {
		return LifecycleTOTPStartView{}, err
	}
	var secret [32]byte
	if _, err = io.ReadFull(rand.Reader, secret[:]); err != nil || secret == [32]byte{} {
		return LifecycleTOTPStartView{}, ErrLifecycleHTTPUnavailable
	}
	defer clear(secret[:])
	environmentID, err := application.pool.ResolveEnvironmentID(ctx)
	if err != nil {
		return LifecycleTOTPStartView{}, ErrLifecycleHTTPUnavailable
	}
	envelope, err := application.totpSealer.SealTOTP(ctx, environmentID, claims.Tenant(),
		claims.Subject(), authenticatorID, secret)
	if err != nil {
		return LifecycleTOTPStartView{}, ErrLifecycleHTTPUnavailable
	}
	requestSHA, _ := lifecycleRequestSHA(struct{ Label string }{input.DisplayLabel})
	err = application.inCommand(ctx, claims, "iam.totp.enrollment.start", "update",
		func(scoped context.Context, scope platformtx.TenantScope, permissionID uuid.UUID) error {
			_, executeErr := application.runtime.service.ReserveTOTP(scoped, scope,
				lifecycle.TOTPStartCommand{
					Evidence:        lifecycleEvidence(claims, permissionID, operationID, requestSHA),
					AuthenticatorID: authenticatorID, UserID: claims.Subject(), SessionID: claims.Session(),
					DisplayLabel: input.DisplayLabel, KMSKeyID: envelope.KMSKeyID,
					Nonce: envelope.Nonce[:], Ciphertext: envelope.Ciphertext[:],
					AuthenticationTag: envelope.AuthenticationTag[:], AADSHA256: envelope.AADSHA256,
					Challenge: challenge,
				})
			return executeErr
		})
	if err != nil {
		return LifecycleTOTPStartView{}, err
	}
	encoded := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(secret[:])
	return LifecycleTOTPStartView{LifecycleChallengeView: LifecycleChallengeView{
		AuthenticatorID: authenticatorID, ChallengeID: challengeID,
		Challenge: wire, ChallengeKeyID: challenge.HMACKeyID,
		RPID: application.rpID, Origin: application.origin,
		UserVerificationRequired: true}, Secret: encoded}, nil
}

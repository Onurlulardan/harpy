package iam

import (
	"context"
	"errors"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycle"
	"github.com/google/uuid"
)

var (
	ErrLifecycleHTTPInvalid           = errors.New("IAM lifecycle request is invalid")
	ErrLifecycleHTTPInvalidCredential = errors.New("IAM lifecycle credential is invalid")
	ErrLifecycleHTTPForbidden         = errors.New("IAM lifecycle operation is forbidden")
	ErrLifecycleHTTPNotFound          = errors.New("IAM lifecycle resource was not found")
	ErrLifecycleHTTPConflict          = errors.New("IAM lifecycle operation conflicts")
	ErrLifecycleHTTPUnavailable       = errors.New("IAM lifecycle runtime is unavailable")
)

type LifecycleHTTPApplication interface {
	AddIdentifier(context.Context, accesstoken.Claims, LifecycleIdentifierAddInput) (LifecycleMutationView, error)
	StartEmailVerification(context.Context, accesstoken.Claims, LifecycleResourceInput) error
	VerifyIdentifier(context.Context, LifecycleCredentialInput) error
	TombstoneIdentifier(context.Context, accesstoken.Claims, LifecycleRevisionInput) error
	RetireIdentifier(context.Context, accesstoken.Claims, LifecycleRevisionInput) error
	StartPasswordReset(context.Context, LifecyclePasswordResetStartInput) error
	CompletePasswordReset(context.Context, LifecyclePasswordResetCompleteInput) error
	CompleteInitialAdminSetup(context.Context, LifecyclePasswordResetCompleteInput) error
	ChangePassword(context.Context, accesstoken.Claims, LifecyclePasswordChangeInput) error
	StartPortalSetup(context.Context, accesstoken.Claims, LifecycleResourceInput) error
	CompletePortalSetup(context.Context, LifecyclePasswordResetCompleteInput) error
	ReadPortalSetup(context.Context, accesstoken.Claims, uuid.UUID) (lifecycle.PortalSetup, error)
	DecidePortalSetup(context.Context, accesstoken.Claims, LifecyclePortalDecisionInput) error
	StartWebAuthn(context.Context, accesstoken.Claims, LifecycleEnrollmentStartInput) (LifecycleChallengeView, error)
	CompleteWebAuthn(context.Context, accesstoken.Claims, LifecycleWebAuthnCompleteInput) error
	StartTOTP(context.Context, accesstoken.Claims, LifecycleEnrollmentStartInput) (LifecycleTOTPStartView, error)
	CompleteTOTP(context.Context, accesstoken.Claims, LifecycleTOTPCompleteInput) error
	IssueRecoveryCodes(context.Context, accesstoken.Claims, LifecycleRecoveryIssueInput) (LifecycleRecoveryIssueView, error)
	ListRecoveryCodes(context.Context, accesstoken.Claims) ([]lifecycle.RecoveryCode, error)
	ListAuthenticators(context.Context, accesstoken.Claims, LifecyclePageInput) ([]lifecycle.Authenticator, error)
	SuspendAuthenticator(context.Context, accesstoken.Claims, LifecycleRevisionInput) error
	RetireAuthenticator(context.Context, accesstoken.Claims, LifecycleRevisionInput) error
	ListSessions(context.Context, accesstoken.Claims, LifecyclePageInput) ([]lifecycle.Session, error)
	RevokeSession(context.Context, accesstoken.Claims, LifecycleRevisionInput) error
	RevokeAllSessions(context.Context, accesstoken.Claims, string) error
}

type LifecycleIdentifierAddInput struct{ IdempotencyKey, Email string }
type LifecycleResourceInput struct {
	ID             uuid.UUID
	IdempotencyKey string
}
type LifecycleCredentialInput struct {
	TenantSlug string
	Token      []byte
}
type LifecycleRevisionInput struct {
	ID               uuid.UUID
	ExpectedRevision int64
	IdempotencyKey   string
}
type LifecyclePasswordResetStartInput struct {
	TenantSlug, IdentifierKind, Identifier, IdempotencyKey string
}
type LifecyclePasswordResetCompleteInput struct {
	TenantSlug      string
	Token, Password []byte
}
type LifecyclePasswordChangeInput struct {
	IdempotencyKey               string
	ExpectedRevision             int64
	CurrentPassword, NewPassword []byte
}
type LifecyclePortalDecisionInput struct {
	LifecycleRevisionInput
	Decision string
}
type LifecycleEnrollmentStartInput struct{ IdempotencyKey, DisplayLabel string }
type LifecycleWebAuthnCompleteInput struct {
	IdempotencyKey, ChallengeKeyID string
	AuthenticatorID, ChallengeID   uuid.UUID
	ExpectedAuthenticatorRevision  int64
	Challenge, Credential          []byte
}
type LifecycleTOTPCompleteInput struct {
	IdempotencyKey, ChallengeKeyID                           string
	AuthenticatorID, ChallengeID                             uuid.UUID
	ExpectedAuthenticatorRevision, ExpectedChallengeRevision int64
	Challenge, Code                                          []byte
}
type LifecycleRecoveryIssueInput struct{ IdempotencyKey, DisplayLabel string }
type LifecyclePageInput struct {
	Limit     int
	AfterTime time.Time
	AfterID   uuid.UUID
}

type LifecycleMutationView struct {
	ID       uuid.UUID `json:"id"`
	Revision int64     `json:"revision"`
}
type LifecycleChallengeView struct {
	AuthenticatorID          uuid.UUID `json:"authenticator_id"`
	ChallengeID              uuid.UUID `json:"challenge_id"`
	Challenge                string    `json:"challenge"`
	ChallengeKeyID           string    `json:"challenge_key_id"`
	RPID                     string    `json:"rp_id"`
	Origin                   string    `json:"origin"`
	UserVerificationRequired bool      `json:"user_verification_required"`
}
type LifecycleTOTPStartView struct {
	LifecycleChallengeView
	Secret string `json:"secret"`
}
type LifecycleRecoveryIssueView struct {
	AuthenticatorID uuid.UUID `json:"authenticator_id"`
	Codes           []string  `json:"codes"`
}

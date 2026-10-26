package lifecycle

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type Store interface {
	LookupCredential(context.Context, platformtx.CredentialTenantScope, CredentialLookup) (CredentialRecord, error)
	EmailDeliveryTarget(context.Context, platformtx.TenantScope, EmailDeliveryTargetQuery) (DeliveryTarget, error)
	PasswordDeliveryTarget(context.Context, platformtx.CredentialTenantScope, PasswordDeliveryTargetQuery) (DeliveryTarget, error)
	PortalDeliveryTarget(context.Context, platformtx.TenantScope, PortalDeliveryTargetQuery) (DeliveryTarget, error)
	AddIdentifier(context.Context, platformtx.TenantScope, IdentifierAddCommand) (MutationResult, error)
	VerifyIdentifier(context.Context, platformtx.CredentialTenantScope, IdentifierVerifyCommand) (MutationResult, error)
	TombstoneIdentifier(context.Context, platformtx.TenantScope, RevisionCommand) (MutationResult, error)
	ArchiveIdentifier(context.Context, platformtx.TenantScope, RevisionCommand) (MutationResult, error)
	IssueEmailVerification(context.Context, platformtx.TenantScope, EmailVerificationStartCommand) (MutationResult, error)
	StartPasswordReset(context.Context, platformtx.CredentialTenantScope, PasswordResetStartCommand) (MutationResult, error)
	CompletePasswordReset(context.Context, platformtx.CredentialTenantScope, PasswordResetCompleteCommand) (MutationResult, error)
	AcceptInitialAdminSetup(context.Context, platformtx.CredentialTenantScope, InitialAdminSetupAcceptCommand) (MutationResult, error)
	CreatePortalSetup(context.Context, platformtx.TenantScope, PortalSetupStartCommand) (MutationResult, error)
	RedeemPortalSetup(context.Context, platformtx.CredentialTenantScope, PortalSetupCompleteCommand) (MutationResult, error)
	ResolvePortalSetup(context.Context, platformtx.TenantScope, PortalSetupQuery) (PortalSetup, error)
	DecidePortalApproval(context.Context, platformtx.TenantScope, PortalDecisionCommand) (MutationResult, error)
	PasswordCredentialLookup(context.Context, platformtx.TenantScope, PasswordChangeCommand) (PasswordCredential, error)
	ChangePassword(context.Context, platformtx.TenantScope, PasswordChangeCommand) (MutationResult, error)
	ListAuthenticators(context.Context, platformtx.TenantScope, AuthenticatorListQuery) ([]Authenticator, error)
	SuspendAuthenticator(context.Context, platformtx.TenantScope, RevisionCommand) (MutationResult, error)
	RetireAuthenticator(context.Context, platformtx.TenantScope, RevisionCommand) (MutationResult, error)
	EnableWebAuthn(context.Context, platformtx.TenantScope, WebAuthnStartCommand) (MutationResult, error)
	FinalizeWebAuthn(context.Context, platformtx.TenantScope, WebAuthnCompleteCommand) (MutationResult, error)
	ReserveTOTP(context.Context, platformtx.TenantScope, TOTPStartCommand) (MutationResult, error)
	TOTPEnrollmentLookup(context.Context, platformtx.TenantScope, TOTPReadQuery) (TOTPEnrollment, error)
	ActivateTOTP(context.Context, platformtx.TenantScope, TOTPCompleteCommand) (MutationResult, error)
	RotateRecoveryCodes(context.Context, platformtx.TenantScope, RecoveryIssueCommand) (MutationResult, error)
	ReadRecoveryCodes(context.Context, platformtx.TenantScope, RecoveryListQuery) ([]RecoveryCode, error)
	GetSessions(context.Context, platformtx.TenantScope, SessionListQuery) ([]Session, error)
	RevokeSession(context.Context, platformtx.TenantScope, RevisionCommand) (MutationResult, error)
	PurgeSessions(context.Context, platformtx.TenantScope, SessionRevokeAllCommand) (MutationResult, error)
}

type Effects interface {
	WriteTenantChange(context.Context, platformtx.TenantScope, Change) error
	WriteCredentialChange(context.Context, platformtx.CredentialTenantScope, Change) error
}

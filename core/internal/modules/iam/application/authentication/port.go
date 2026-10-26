package authentication

import (
	"context"
)

type Port interface {
	ReauthenticationCredential(
		context.Context,
		CredentialScope,
		ReauthenticationLookup,
	) (ReauthenticationCredential, error)
	CompleteReauthentication(
		context.Context,
		CredentialScope,
		ReauthenticationCompletion,
	) (ReauthenticationCompleted, error)
	ReauthenticationRejection(
		context.Context,
		CredentialScope,
		ReauthenticationRejection,
	) (ReauthenticationRejected, error)
	RecoveryCredential(
		context.Context,
		CredentialScope,
		RecoveryLookup,
	) (RecoveryCredential, error)
	ConsumeRecovery(
		context.Context,
		CredentialScope,
		RecoveryConsumption,
	) (RecoveryConsumed, error)
	RevokeSession(
		context.Context,
		CredentialScope,
		LogoutMutation,
	) (LogoutResult, error)
	StartReauthentication(
		context.Context,
		CredentialScope,
		ChallengeCreation,
	) (ChallengeCreated, error)
	LoginCredential(
		context.Context,
		CredentialScope,
	) (LoginCredential, error)
	UsePassword(
		context.Context,
		CredentialScope,
		PasswordUse,
	) (PasswordUsed, error)
	Issue(
		context.Context,
		CredentialScope,
		SessionIssue,
	) (SessionIssued, error)
	RefreshCredential(
		context.Context,
		CredentialScope,
	) (RefreshCredential, error)
	Rotate(
		context.Context,
		CredentialScope,
		RefreshRotation,
	) (RefreshRotated, error)
}

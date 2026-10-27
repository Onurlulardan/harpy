package authentication

import "context"

type InitialPasswordPort interface {
	InitialPasswordCredential(
		context.Context, CredentialScope,
	) (InitialPasswordCredential, error)
	ClaimInitialPassword(
		context.Context, CredentialScope, InitialPasswordClaim,
	) (InitialPasswordClaimed, error)
	InitialPasswordContinuation(
		context.Context, CredentialScope, [16]byte,
	) (InitialPasswordContinuation, error)
	RejectInitialPassword(
		context.Context, CredentialScope, InitialPasswordReject,
	) (InitialPasswordRejected, error)
	FinalizeInitialPassword(
		context.Context, CredentialScope, InitialPasswordFinalize,
	) (InitialPasswordFinalized, error)
	InitialSession(
		context.Context, CredentialScope, SessionIssue,
	) (SessionIssued, error)
}

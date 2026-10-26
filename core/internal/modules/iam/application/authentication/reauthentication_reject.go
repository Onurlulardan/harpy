package authentication

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (authenticator Authenticator) rejectReauthentication(
	ctx context.Context,
	scope platformtx.CredentialTenantScope,
	credential ReauthenticationCredential,
	quarantine, expire bool,
) error {
	rejection := ReauthenticationRejection{
		ChallengeID:          credential.ChallengeID,
		ExpectedRevision:     credential.ChallengeRevision,
		ExpectedAttemptCount: credential.AttemptCount,
		Expire:               expire,
	}
	if quarantine {
		rejection.AuthenticatorID = credential.AuthenticatorID
		rejection.ExpectedAuthenticatorRevision = credential.AuthenticatorRevision
		rejection.QuarantineAuthenticator = true
	}
	result, err := authenticator.port.ReauthenticationRejection(
		ctx, scope, rejection,
	)
	if err != nil || result.validate(rejection) != nil {
		return ErrAuthenticationStorage
	}
	return nil
}

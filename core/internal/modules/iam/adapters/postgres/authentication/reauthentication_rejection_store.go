package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authentication"
	"github.com/google/uuid"
)

func (store *store) ReauthenticationRejection(
	ctx context.Context,
	scope authentication.CredentialScope,
	rejection authentication.ReauthenticationRejection,
) (authentication.ReauthenticationRejected, error) {
	bound, err := store.bindReauthentication(ctx, scope)
	if err != nil {
		return authentication.ReauthenticationRejected{}, err
	}
	var authenticatorID any
	var authenticatorRevision any
	if rejection.AuthenticatorID != uuid.Nil {
		authenticatorID = rejection.AuthenticatorID
		authenticatorRevision = rejection.ExpectedAuthenticatorRevision
	}
	var result authentication.ReauthenticationRejected
	err = bound.database.QueryRow(ctx, rejectReauthenticationSQL,
		rejection.ChallengeID, rejection.ExpectedRevision,
		rejection.ExpectedAttemptCount, authenticatorID, authenticatorRevision,
		rejection.QuarantineAuthenticator, rejection.Expire,
	).Scan(&result.State, &result.Revision, &result.AttemptCount)
	if err != nil {
		return authentication.ReauthenticationRejected{}, errAuthenticationStorage
	}
	return result, nil
}

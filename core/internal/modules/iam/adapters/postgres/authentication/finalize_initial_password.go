package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authentication"
)

func (store *store) FinalizeInitialPassword(
	ctx context.Context,
	scope authentication.CredentialScope,
	command authentication.InitialPasswordFinalize,
) (authentication.InitialPasswordFinalized, error) {
	bound, err := store.bindInitialRoute(ctx, scope)
	if err != nil {
		return authentication.InitialPasswordFinalized{}, err
	}
	secretHMAC, csrfHMAC := command.SecretHMAC, command.CSRFHMAC
	salt, tag, requestSHA := command.Salt, command.Tag, command.RequestSHA256
	var result authentication.InitialPasswordFinalized
	err = bound.database.QueryRow(ctx, initialPasswordCompleteSQL, scope.TenantID(),
		command.Selector[:], secretHMAC[:], csrfHMAC[:],
		command.ExpectedContinuationRevision, command.UserID, command.CredentialID,
		command.ExpectedCredentialVersion, command.ExpectedCredentialRevision,
		salt[:], tag[:], requestSHA[:],
	).Scan(&result.UserID, &result.AccessModeID, &result.UserRevision,
		&result.TokenVersion, &result.AuthorizationVersion, &result.AccessModeRevision,
		&result.CredentialVersion, &result.CredentialRevision,
		&result.ContinuationRevision, &result.InstallationRevision)
	clear(secretHMAC[:])
	clear(csrfHMAC[:])
	clear(salt[:])
	clear(tag[:])
	if err != nil {
		return authentication.InitialPasswordFinalized{}, errAuthenticationStorage
	}
	return result, nil
}

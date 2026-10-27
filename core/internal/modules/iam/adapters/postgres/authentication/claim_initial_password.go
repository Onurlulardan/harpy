package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authentication"
)

func (store *store) ClaimInitialPassword(
	ctx context.Context,
	scope authentication.CredentialScope,
	command authentication.InitialPasswordClaim,
) (authentication.InitialPasswordClaimed, error) {
	bound, err := store.bindInitialLogin(ctx, scope)
	if err != nil {
		return authentication.InitialPasswordClaimed{}, err
	}
	selector, secretHMAC, csrfHMAC := command.Selector, command.SecretHMAC, command.CSRFHMAC
	var result authentication.InitialPasswordClaimed
	err = bound.database.QueryRow(ctx, initialPasswordStartSQL, scope.TenantID(),
		command.UserID, command.InstallationOperationID, command.CredentialID,
		command.CredentialVersion, command.CredentialRevision, selector[:],
		command.SecretKeyID, secretHMAC[:], command.CSRFKeyID, csrfHMAC[:],
	).Scan(&result.Revision, &result.IssuedAt, &result.ExpiresAt)
	clear(secretHMAC[:])
	clear(csrfHMAC[:])
	if err != nil {
		return authentication.InitialPasswordClaimed{}, errAuthenticationStorage
	}
	return canonicalInitialPasswordClaimed(result), nil
}

func canonicalInitialPasswordClaimed(
	result authentication.InitialPasswordClaimed,
) authentication.InitialPasswordClaimed {
	result.IssuedAt = result.IssuedAt.UTC()
	result.ExpiresAt = result.ExpiresAt.UTC()
	return result
}

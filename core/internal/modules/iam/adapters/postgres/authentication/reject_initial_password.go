package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authentication"
)

func (store *store) RejectInitialPassword(
	ctx context.Context,
	scope authentication.CredentialScope,
	command authentication.InitialPasswordReject,
) (authentication.InitialPasswordRejected, error) {
	bound, err := store.bindInitialRoute(ctx, scope)
	if err != nil {
		return authentication.InitialPasswordRejected{}, err
	}
	secretHMAC, csrfHMAC := command.SecretHMAC, command.CSRFHMAC
	var result authentication.InitialPasswordRejected
	err = bound.database.QueryRow(ctx, initialPasswordRejectSQL, scope.TenantID(),
		command.Selector[:], secretHMAC[:], csrfHMAC[:], command.ExpectedRevision,
	).Scan(&result.State, &result.PolicyAttempts, &result.Revision)
	clear(secretHMAC[:])
	clear(csrfHMAC[:])
	if err != nil {
		return authentication.InitialPasswordRejected{}, errAuthenticationStorage
	}
	return result, nil
}

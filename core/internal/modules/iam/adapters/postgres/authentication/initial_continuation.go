package postgres

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authentication"
)

func (store *store) InitialPasswordContinuation(
	ctx context.Context,
	scope authentication.CredentialScope,
	selector [16]byte,
) (authentication.InitialPasswordContinuation, error) {
	bound, err := store.bindInitialRoute(ctx, scope)
	if err != nil {
		return authentication.InitialPasswordContinuation{}, err
	}
	var result authentication.InitialPasswordContinuation
	var storedSelector, secretHMAC, csrfHMAC []byte
	var updatedAt time.Time
	var consumedAt *time.Time
	err = bound.database.QueryRow(ctx, initialPasswordContinuationSQL,
		scope.TenantID(), selector[:],
	).Scan(&storedSelector, &result.TenantID, &result.UserID,
		&result.InstallationOperationID, &result.CredentialID,
		&result.CredentialVersion, &result.CredentialRevision,
		&result.SecretKeyID, &secretHMAC, &result.CSRFKeyID, &csrfHMAC,
		&result.State, &result.PolicyAttempts, &result.Revision,
		&result.IssuedAt, &result.ExpiresAt, &updatedAt, &consumedAt)
	if len(storedSelector) == len(result.Selector) {
		copy(result.Selector[:], storedSelector)
	}
	if len(secretHMAC) == len(result.SecretHMAC) {
		copy(result.SecretHMAC[:], secretHMAC)
	}
	if len(csrfHMAC) == len(result.CSRFHMAC) {
		copy(result.CSRFHMAC[:], csrfHMAC)
	}
	clear(storedSelector)
	clear(secretHMAC)
	clear(csrfHMAC)
	if err != nil {
		return authentication.InitialPasswordContinuation{}, errAuthenticationStorage
	}
	return result, nil
}

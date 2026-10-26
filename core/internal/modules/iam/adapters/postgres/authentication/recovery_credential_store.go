package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authentication"
)

func (store *store) RecoveryCredential(
	ctx context.Context,
	scope authentication.CredentialScope,
	lookup authentication.RecoveryLookup,
) (authentication.RecoveryCredential, error) {
	bound, err := store.bindRecovery(ctx, scope)
	if err != nil {
		return authentication.RecoveryCredential{}, err
	}
	binding := scope.LocatorBindingSHA256()
	codeHMAC := lookup.CodeHMAC
	var result authentication.RecoveryCredential
	var storedHMAC []byte
	err = bound.database.QueryRow(ctx, resolveRecoveryCredentialSQL,
		binding[:], lookup.HMACKeyID, codeHMAC[:],
	).Scan(
		&result.CodeID, &result.AuthenticatorID, &result.UserID,
		&result.CodeRevision, &result.AuthenticatorRevision,
		&result.UserRevision, &result.TokenVersion, &result.CodeState,
		&result.AuthenticatorState, &result.UserStatus, &result.IdentityKind,
		&result.HMACKeyID, &storedHMAC,
	)
	if err != nil || len(storedHMAC) != len(result.CodeHMAC) {
		clear(storedHMAC)
		return authentication.RecoveryCredential{}, errAuthenticationStorage
	}
	copy(result.CodeHMAC[:], storedHMAC)
	clear(storedHMAC)
	return result, nil
}

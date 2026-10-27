package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authentication"
)

func (store *store) ConsumeRecovery(
	ctx context.Context,
	scope authentication.CredentialScope,
	consumption authentication.RecoveryConsumption,
) (authentication.RecoveryConsumed, error) {
	bound, err := store.bindRecovery(ctx, scope)
	if err != nil {
		return authentication.RecoveryConsumed{}, err
	}
	binding := scope.LocatorBindingSHA256()
	codeHMAC := consumption.CodeHMAC
	var result authentication.RecoveryConsumed
	err = bound.database.QueryRow(ctx, consumeRecoverySQL,
		binding[:], consumption.HMACKeyID, codeHMAC[:], consumption.CodeID,
		consumption.ExpectedCodeRevision, consumption.AuthenticatorID,
		consumption.ExpectedAuthenticatorRevision, consumption.UserID,
		consumption.ExpectedUserRevision, consumption.ExpectedTokenVersion,
	).Scan(&result.TokenVersion)
	if err != nil {
		return authentication.RecoveryConsumed{}, errAuthenticationStorage
	}
	return result, nil
}

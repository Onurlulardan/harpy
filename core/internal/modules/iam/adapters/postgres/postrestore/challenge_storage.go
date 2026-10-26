package postrestore

import (
	"context"

	application "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/postrestore"
)

func (store *Store) PersistChallenge(
	ctx context.Context, command application.ChallengeCommand,
) (application.ChallengeRecord, error) {
	var result application.ChallengeRecord
	err := store.global(ctx, transactionCommand, func(database database) error {
		return database.QueryRow(ctx, issueChallengeSQL,
			command.OperationID, command.EnvironmentID, command.RestoreOperationID,
			command.Challenge, command.ContextSHA256[:], command.RPID, command.Origin,
			command.ExpiresAt, command.OccurredAt,
		).Scan(&result.OperationID, &result.TenantID, &result.OwnerUserID,
			&result.AnchorID, &result.Challenge, &result.CredentialID,
			&result.PublicKeyCOSE, &result.RPID, &result.SignCount,
			&result.ExpiresAt, &result.Replayed)
	})
	return result, err
}

package postrestore

import (
	"context"
	"time"

	application "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/postrestore"
	"github.com/google/uuid"
)

func (store *Store) ChallengeContext(
	ctx context.Context, operationID, environmentID, restoreOperationID uuid.UUID,
	origin string, occurredAt time.Time,
) (application.ChallengeContext, error) {
	var result application.ChallengeContext
	var contextDigest []byte
	err := store.global(ctx, snapshotQuery, func(database database) error {
		return database.QueryRow(ctx, challengeContextSQL, operationID, environmentID,
			restoreOperationID, origin, occurredAt).Scan(
			&result.TenantID, &result.OwnerUserID, &result.AnchorID, &result.Challenge,
			&result.CredentialID, &result.PublicKeyCOSE, &result.RPID, &result.SignCount,
			&contextDigest)
	})
	if err == nil && len(contextDigest) == 32 {
		copy(result.ContextSHA256[:], contextDigest)
	} else if err == nil {
		return application.ChallengeContext{}, application.ErrUnavailable
	}
	if err != nil {
		return application.ChallengeContext{}, err
	}
	var digest []byte
	err = store.tenant(ctx, snapshotQuery, result.TenantID,
		func(database database) error {
			return database.QueryRow(ctx, ownerContextSQL, result.OwnerUserID).Scan(
				&result.Identifier.ID, &result.Identifier.NormalizedValue,
				&digest, &result.Identifier.Revision)
		})
	if err == nil && len(digest) == 32 {
		copy(result.Identifier.NormalizedSHA256[:], digest)
	} else if err == nil {
		return application.ChallengeContext{}, application.ErrUnavailable
	}
	return result, err
}

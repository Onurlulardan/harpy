package postgres

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycle"
)

func (store *store) ReadRecoveryCodes(ctx context.Context,
	scope tenantScope,
	query lifecycle.RecoveryListQuery) ([]lifecycle.RecoveryCode, error) {
	if !snapshotUsable(ctx, scope) || scope.ActorID() != query.ActorID {
		return nil, lifecycle.ErrInvalid
	}
	bound, err := store.bindTenant(scope)
	if err != nil {
		return nil, err
	}
	rows, err := bound.database.Query(ctx, recoveryCodesListSQL, query.UserID,
		query.ActorID, query.PermissionID, query.ExpectedTokenVersion,
		query.ExpectedAuthorizationVersion)
	if err != nil {
		return nil, mapError(err)
	}
	defer rows.Close()
	var result []lifecycle.RecoveryCode
	for rows.Next() {
		var value lifecycle.RecoveryCode
		var consumedAt, revokedAt *time.Time
		if err := rows.Scan(&value.AuthenticatorID, &value.AuthenticatorRevision,
			&value.CodeID, &value.Position, &value.State, &value.CreatedAt,
			&consumedAt, &revokedAt); err != nil {
			return nil, lifecycle.ErrUnavailable
		}
		if consumedAt != nil {
			value.ConsumedAt = consumedAt.UTC()
		}
		if revokedAt != nil {
			value.RevokedAt = revokedAt.UTC()
		}
		result = append(result, value)
	}
	if rows.Err() != nil {
		return nil, lifecycle.ErrUnavailable
	}
	return result, nil
}

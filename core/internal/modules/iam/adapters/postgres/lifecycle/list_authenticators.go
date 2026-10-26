package postgres

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycle"
)

func (store *store) ListAuthenticators(ctx context.Context,
	scope tenantScope,
	query lifecycle.AuthenticatorListQuery) ([]lifecycle.Authenticator, error) {
	if !snapshotUsable(ctx, scope) || scope.ActorID() != query.ActorID {
		return nil, lifecycle.ErrInvalid
	}
	bound, err := store.bindTenant(scope)
	if err != nil {
		return nil, err
	}
	rows, err := bound.database.Query(ctx, authenticatorListSQL, query.UserID,
		query.ActorID, query.PermissionID, query.ExpectedTokenVersion,
		query.ExpectedAuthorizationVersion, query.Limit,
		nullableTime(query.AfterUpdatedAt), nullableUUID(query.AfterID))
	if err != nil {
		return nil, mapError(err)
	}
	defer rows.Close()
	result := make([]lifecycle.Authenticator, 0, query.Limit)
	for rows.Next() {
		var value lifecycle.Authenticator
		var activatedAt, lastUsedAt *time.Time
		if err := rows.Scan(&value.ID, &value.UserID, &value.Kind, &value.DisplayLabel,
			&value.State, &value.Revision, &value.CreatedAt, &value.UpdatedAt,
			&activatedAt, &lastUsedAt); err != nil {
			return nil, lifecycle.ErrUnavailable
		}
		if activatedAt != nil {
			value.ActivatedAt = activatedAt.UTC()
		}
		if lastUsedAt != nil {
			value.LastUsedAt = lastUsedAt.UTC()
		}
		result = append(result, value)
	}
	if rows.Err() != nil {
		return nil, lifecycle.ErrUnavailable
	}
	return result, nil
}

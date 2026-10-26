package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycle"
	"github.com/google/uuid"
)

func (store *store) GetSessions(ctx context.Context, scope tenantScope,
	query lifecycle.SessionListQuery) ([]lifecycle.Session, error) {
	if !snapshotUsable(ctx, scope) || scope.ActorID() != query.ActorID {
		return nil, lifecycle.ErrInvalid
	}
	bound, err := store.bindTenant(scope)
	if err != nil {
		return nil, err
	}
	rows, err := bound.database.Query(ctx, sessionListSQL, query.UserID,
		query.ActorID, query.PermissionID, query.ExpectedTokenVersion,
		query.ExpectedAuthorizationVersion, query.Limit,
		nullableTime(query.AfterLastSeenAt), nullableUUID(query.AfterID))
	if err != nil {
		return nil, mapError(err)
	}
	defer rows.Close()
	result := make([]lifecycle.Session, 0, query.Limit)
	for rows.Next() {
		var value lifecycle.Session
		var portalBindingID *uuid.UUID
		if err := rows.Scan(&value.ID, &value.AccessMode, &portalBindingID,
			&value.DeviceKind, &value.BrowserFamily, &value.OSFamily, &value.Revision,
			&value.CreatedAt, &value.LastSeenAt, &value.AbsoluteExpiry); err != nil {
			return nil, lifecycle.ErrUnavailable
		}
		if portalBindingID != nil {
			value.PortalBindingID = *portalBindingID
		}
		result = append(result, value)
	}
	if rows.Err() != nil {
		return nil, lifecycle.ErrUnavailable
	}
	return result, nil
}

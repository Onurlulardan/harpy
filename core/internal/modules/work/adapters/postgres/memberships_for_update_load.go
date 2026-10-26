package postgres

import (
	"context"
	_ "embed"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/google/uuid"
)

//go:embed sql/tenant/list_memberships_for_update.sql
var listMembershipsForUpdateSQL string

func (store *Store) LoadMembershipsForUpdate(ctx context.Context, boardID domain.ID) ([]domain.Membership, error) {
	if !boardID.Valid() {
		return nil, application.ErrInvalid
	}
	if err := store.check(ctx); err != nil {
		return nil, err
	}
	rows, err := store.database.Query(ctx, listMembershipsForUpdateSQL, boardID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]domain.Membership, 0)
	for rows.Next() {
		var id, storedBoardID, principalID uuid.UUID
		var kind, role string
		var active bool
		var version int64
		if err = rows.Scan(&id, &storedBoardID, &kind, &principalID, &role, &active, &version); err != nil {
			return nil, err
		}
		item := domain.Membership{ID: domain.ID(id), BoardID: domain.ID(storedBoardID),
			Principal: domain.Principal{Kind: domain.PrincipalKind(kind), ID: domain.ID(principalID)},
			Role:      domain.BoardRole(role), Active: active, RowVersion: version}
		if !item.Valid() {
			return nil, application.ErrUnavailable
		}
		items = append(items, item)
	}
	if err = rows.Err(); err != nil || len(items) > 10_000 {
		return nil, application.ErrUnavailable
	}
	return items, nil
}

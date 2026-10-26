package postgres

import (
	"context"
	_ "embed"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
)

//go:embed sql/tenant/upsert_membership.sql
var upsertMembershipSQL string

func (store *Store) UpsertMembership(ctx context.Context, item domain.Membership) error {
	if !item.Valid() || !item.Active {
		return domain.ErrVersionConflict
	}
	return executeOne(ctx, store, upsertMembershipSQL, item.ID, item.BoardID,
		item.Principal.Kind, item.Principal.ID, item.Role, item.RowVersion)
}

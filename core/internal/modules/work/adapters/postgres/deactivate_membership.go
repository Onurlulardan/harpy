package postgres

import (
	"context"
	_ "embed"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
)

//go:embed sql/tenant/deactivate_membership.sql
var deactivateMembershipSQL string

func (store *Store) DeactivateMembership(ctx context.Context, id domain.ID, expected int64) error {
	if !id.Valid() || expected < 1 {
		return domain.ErrVersionConflict
	}
	return executeOne(ctx, store, deactivateMembershipSQL, id, expected)
}

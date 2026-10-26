package postgres

import (
	"context"
	"crypto/sha256"
	_ "embed"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
)

//go:embed sql/tenant/mark_projection_repair.sql
var markProjectionRepairSQL string

func (store *Store) FlagRepair(ctx context.Context, owner domain.SourceOwner, kind string, id domain.ID, sequence int64) error {
	if owner == domain.SourceWork || !owner.Valid() || kind == "" || !id.Valid() || sequence < 1 {
		return application.ErrInvalid
	}
	hash := sha256.Sum256([]byte(owner + domain.SourceOwner(":"+kind+":"+id.String())))
	return executeOne(ctx, store, markProjectionRepairSQL, owner, kind, id, sequence, hash[:])
}

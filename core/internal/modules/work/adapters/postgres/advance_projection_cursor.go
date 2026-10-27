package postgres

import (
	"context"
	"crypto/sha256"
	_ "embed"
	"strconv"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
)

//go:embed sql/tenant/advance_projection_cursor.sql
var advanceProjectionCursorSQL string

func (store *Store) AdvanceCursor(ctx context.Context, owner domain.SourceOwner, kind string, id domain.ID, sequence int64) error {
	if owner == domain.SourceWork || !owner.Valid() || kind == "" || !id.Valid() || sequence < 1 {
		return application.ErrInvalid
	}
	hash := sha256.Sum256([]byte(string(owner) + ":" + kind + ":" + id.String() + ":" + versionText(sequence)))
	return executeOne(ctx, store, advanceProjectionCursorSQL, owner, kind, id, sequence, hash[:])
}

func versionText(value int64) string { return strconv.FormatInt(value, 10) }

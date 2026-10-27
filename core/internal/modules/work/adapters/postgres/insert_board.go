package postgres

import (
	"context"
	_ "embed"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
)

//go:embed sql/tenant/insert_board.sql
var insertBoardSQL string

func (store *Store) InsertBoard(ctx context.Context, board domain.Board) error {
	if !board.Valid() {
		return domain.ErrVersionConflict
	}
	return executeOne(ctx, store, insertBoardSQL, board.ID, board.CreatorID, board.Name,
		board.Description, board.Lifecycle, board.RowVersion, board.CreatedAt)
}

package postgres

import (
	"context"
	_ "embed"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
)

//go:embed sql/tenant/update_board.sql
var updateBoardSQL string

func (store *Store) PersistBoard(ctx context.Context, board domain.Board, expected int64) error {
	if !board.Valid() || expected < 1 || board.RowVersion != expected+1 {
		return domain.ErrVersionConflict
	}
	return executeOne(ctx, store, updateBoardSQL, board.ID, board.Name, board.Description,
		board.Lifecycle, board.RowVersion, expected)
}

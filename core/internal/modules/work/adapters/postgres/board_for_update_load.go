package postgres

import (
	"context"
	_ "embed"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/google/uuid"
)

//go:embed sql/tenant/get_board_for_update.sql
var getBoardForUpdateSQL string

func (store *Store) LoadBoardForUpdate(ctx context.Context, id domain.ID) (domain.Board, error) {
	if !id.Valid() {
		return domain.Board{}, application.ErrInvalid
	}
	if err := store.check(ctx); err != nil {
		return domain.Board{}, err
	}
	var boardID, creatorID uuid.UUID
	var name, description, lifecycle string
	var version int64
	var createdAt time.Time
	err := store.database.QueryRow(ctx, getBoardForUpdateSQL, id).Scan(
		&boardID, &creatorID, &name, &description, &lifecycle, &version, &createdAt)
	if err != nil {
		return domain.Board{}, err
	}
	board := domain.Board{ID: domain.ID(boardID), CreatorID: domain.ID(creatorID), Name: name,
		Description: description, Lifecycle: domain.Lifecycle(lifecycle),
		RowVersion: version, CreatedAt: createdAt.UTC()}
	if !board.Valid() {
		return domain.Board{}, application.ErrUnavailable
	}
	return board, nil
}

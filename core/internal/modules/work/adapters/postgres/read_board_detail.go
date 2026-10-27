package postgres

import (
	"context"
	_ "embed"
	"encoding/json"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
	"github.com/google/uuid"
)

//go:embed sql/tenant/read_visible_board_detail.sql
var readVisibleBoardDetailSQL string

func (reader *Reader) ReadBoardDetail(ctx context.Context, local ports.Scope,
	boardID domain.ID,
) (json.RawMessage, error) {
	bridge, ok := local.(*scopeBridge)
	if reader == nil || !ok || !scopeUsable(ctx, bridge) || !boardID.Valid() {
		return nil, application.ErrInvalid
	}
	snapshot, err := reader.principals.Resolve(ctx, local)
	if err != nil {
		return nil, application.ErrForbidden
	}
	kinds, ids := principalArrays(snapshot.Principals)
	database, err := reader.access.bind(ctx, bridge)
	if err != nil {
		return nil, application.ErrUnavailable
	}
	var document []byte
	err = database.QueryRow(ctx, readVisibleBoardDetailSQL,
		uuid.UUID(boardID), kinds, ids).Scan(&document)
	if err != nil || !json.Valid(document) {
		return nil, application.ErrUnavailable
	}
	return append(json.RawMessage(nil), document...), nil
}

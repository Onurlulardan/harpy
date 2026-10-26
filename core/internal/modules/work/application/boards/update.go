package boards

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

type UpdateBoardInput struct{ Name, Description string }

func (service *Service) UpdateBoard(ctx context.Context, scope ports.Scope, meta ports.Mutation,
	boardID domain.ID, input UpdateBoardInput,
) (ports.OperationReceipt, error) {
	name := strings.TrimSpace(input.Name)
	if !boardID.Valid() || name == "" || len(name) > 200 ||
		len(input.Description) > 16_000 || meta.ExpectedVersion < 1 {
		return ports.OperationReceipt{}, ports.ErrInvalid
	}
	hash := ports.Fingerprint(boardID.String(), name, input.Description,
		ports.VersionString(meta.ExpectedVersion))
	return service.mutator.Mutate(ctx, scope, "work.board.update", boardID, meta, hash,
		func(ctx context.Context, stores ports.Stores) (ports.OperationReceipt, []byte, error) {
			board, err := stores.Boards().LoadBoardForUpdate(ctx, boardID)
			if err != nil {
				return ports.OperationReceipt{}, nil, err
			}
			if board.RowVersion != meta.ExpectedVersion || board.Lifecycle != domain.LifecycleActive {
				return ports.OperationReceipt{}, nil, ports.ErrConflict
			}
			board.Name, board.Description, board.RowVersion = name, input.Description, board.RowVersion+1
			if !board.Valid() {
				return ports.OperationReceipt{}, nil, ports.ErrInvalid
			}
			if err = stores.Boards().PersistBoard(ctx, board, meta.ExpectedVersion); err != nil {
				return ports.OperationReceipt{}, nil, err
			}
			payload, _ := json.Marshal(struct {
				BoardID string
				Version int64
			}{
				boardID.String(), board.RowVersion})
			return ports.OperationReceipt{AggregateID: boardID, RowVersion: board.RowVersion}, payload, nil
		})
}

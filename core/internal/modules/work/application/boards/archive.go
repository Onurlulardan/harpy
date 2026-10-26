package boards

import (
	"context"
	"encoding/json"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func (service *Service) ArchiveBoard(ctx context.Context, scope ports.Scope,
	meta ports.Mutation, boardID domain.ID) (ports.OperationReceipt, error) {
	return service.writeLifecycle(ctx, scope, meta, boardID, domain.LifecycleArchived,
		"work.board.archive")
}

func (service *Service) writeLifecycle(ctx context.Context, scope ports.Scope,
	meta ports.Mutation, boardID domain.ID, lifecycle domain.Lifecycle,
	authority string) (ports.OperationReceipt, error) {
	if !boardID.Valid() || meta.ExpectedVersion < 1 {
		return ports.OperationReceipt{}, ports.ErrInvalid
	}
	hash := ports.Fingerprint(boardID.String(), string(lifecycle),
		ports.VersionString(meta.ExpectedVersion))
	return service.mutator.Mutate(ctx, scope, authority, boardID, meta, hash,
		func(ctx context.Context, stores ports.Stores) (ports.OperationReceipt, []byte, error) {
			board, err := stores.Boards().LoadBoardForUpdate(ctx, boardID)
			if err != nil {
				return ports.OperationReceipt{}, nil, err
			}
			if board.RowVersion != meta.ExpectedVersion || board.Lifecycle == lifecycle {
				return ports.OperationReceipt{}, nil, ports.ErrConflict
			}
			board.Lifecycle, board.RowVersion = lifecycle, board.RowVersion+1
			if err = stores.Boards().PersistBoard(ctx, board, meta.ExpectedVersion); err != nil {
				return ports.OperationReceipt{}, nil, err
			}
			payload, _ := json.Marshal(struct {
				BoardID, Lifecycle string
				Version            int64
			}{
				boardID.String(), string(lifecycle), board.RowVersion})
			return ports.OperationReceipt{AggregateID: boardID, RowVersion: board.RowVersion}, payload, nil
		})
}

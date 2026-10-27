package columns

import (
	"context"
	"encoding/json"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

type ReorderInput struct {
	BoardID   domain.ID
	ColumnIDs []domain.ID
}

func (service *Service) ReorderColumns(ctx context.Context, scope ports.Scope,
	meta ports.Mutation, input ReorderInput,
) (ports.OperationReceipt, error) {
	if !input.BoardID.Valid() || len(input.ColumnIDs) < 1 ||
		len(input.ColumnIDs) > domain.MaxBoardColumns || meta.ExpectedVersion < 1 {
		return ports.OperationReceipt{}, ports.ErrInvalid
	}
	parts := []string{input.BoardID.String(), ports.VersionString(meta.ExpectedVersion)}
	seen := map[domain.ID]bool{}
	for _, id := range input.ColumnIDs {
		if !id.Valid() || seen[id] {
			return ports.OperationReceipt{}, ports.ErrInvalid
		}
		seen[id] = true
		parts = append(parts, id.String())
	}
	return service.mutator.Mutate(ctx, scope, "work.column.update", input.BoardID,
		meta, ports.Fingerprint(parts...), func(ctx context.Context, stores ports.Stores) (
			ports.OperationReceipt, []byte, error,
		) {
			board, err := stores.Boards().LoadBoardForUpdate(ctx, input.BoardID)
			if err != nil || board.RowVersion != meta.ExpectedVersion ||
				board.Lifecycle != domain.LifecycleActive {
				return ports.OperationReceipt{}, nil, ports.ErrConflict
			}
			stored, err := stores.Boards().LoadColumnsForUpdate(ctx, input.BoardID)
			if err != nil || len(stored) != len(input.ColumnIDs) {
				return ports.OperationReceipt{}, nil, ports.ErrConflict
			}
			byID := make(map[domain.ID]domain.Column, len(stored))
			for _, column := range stored {
				byID[column.ID] = column
			}
			for index, id := range input.ColumnIDs {
				column, found := byID[id]
				if !found {
					return ports.OperationReceipt{}, nil, ports.ErrConflict
				}
				position := int64(index+1) * domain.SparseGap
				if column.Position != position {
					old := column.RowVersion
					column.Position, column.RowVersion = position, old+1
					if err = stores.Boards().PersistColumn(ctx, column, old); err != nil {
						return ports.OperationReceipt{}, nil, err
					}
				}
			}
			old := board.RowVersion
			board.RowVersion++
			if err = stores.Boards().PersistBoard(ctx, board, old); err != nil {
				return ports.OperationReceipt{}, nil, err
			}
			payload, _ := json.Marshal(struct {
				BoardID string   `json:"board_id"`
				Version int64    `json:"version"`
				Order   []string `json:"column_ids"`
			}{board.ID.String(), board.RowVersion, parts[2:]})
			return ports.OperationReceipt{AggregateID: board.ID,
				RowVersion: board.RowVersion}, payload, nil
		})
}

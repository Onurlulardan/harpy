package columns

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

type CreateColumnInput struct {
	BoardID       domain.ID
	Name, Color   string
	Status        domain.TaskStatus
	Before, After *domain.ID
}

func (service *Service) CreateColumn(
	ctx context.Context, scope ports.Scope, meta ports.Mutation, input CreateColumnInput,
) (ports.OperationReceipt, error) {
	input.Name = strings.TrimSpace(input.Name)
	if !input.BoardID.Valid() || input.Name == "" || len(input.Name) > 200 ||
		!domain.ValidColor(input.Color) || !input.Status.Valid() || meta.ExpectedVersion < 1 {
		return ports.OperationReceipt{}, ports.ErrInvalid
	}
	hash := ports.Fingerprint(input.BoardID.String(), input.Name, input.Color, string(input.Status),
		ports.IDString(input.Before), ports.IDString(input.After), ports.VersionString(meta.ExpectedVersion))
	return service.mutator.Mutate(ctx, scope, "work.column.create", input.BoardID, meta, hash,
		func(ctx context.Context, stores ports.Stores) (ports.OperationReceipt, []byte, error) {
			board, err := stores.Boards().LoadBoardForUpdate(ctx, input.BoardID)
			if err != nil {
				return ports.OperationReceipt{}, nil, err
			}
			if board.RowVersion != meta.ExpectedVersion || board.Lifecycle != domain.LifecycleActive {
				return ports.OperationReceipt{}, nil, ports.ErrConflict
			}
			columns, err := stores.Boards().LoadColumnsForUpdate(ctx, input.BoardID)
			if err != nil || len(columns) >= domain.MaxBoardColumns {
				return ports.OperationReceipt{}, nil, domain.ErrBoardLimit
			}
			ranked := make([]domain.RankedTask, 0, len(columns))
			for _, column := range columns {
				ranked = append(ranked, domain.RankedTask{ID: column.ID, Position: column.Position})
			}
			id, err := service.ids.New(service.clock.Now().UTC())
			if err != nil {
				return ports.OperationReceipt{}, nil, ports.ErrUnavailable
			}
			plan, err := domain.PlanPosition(ranked, id, domain.MoveNeighbors{
				Before: input.Before, After: input.After})
			if err != nil {
				return ports.OperationReceipt{}, nil, err
			}
			for _, item := range plan.Rebalanced {
				for _, column := range columns {
					if column.ID == item.ID {
						old := column.RowVersion
						column.Position, column.RowVersion = item.Position, old+1
						if err = stores.Boards().PersistColumn(ctx, column, old); err != nil {
							return ports.OperationReceipt{}, nil, err
						}
					}
				}
			}
			column := domain.Column{ID: id, BoardID: input.BoardID, Name: input.Name,
				Color: input.Color, Status: input.Status, Position: plan.Position, RowVersion: 1}
			if !column.Valid() {
				return ports.OperationReceipt{}, nil, ports.ErrInvalid
			}
			if err = stores.Boards().InsertColumn(ctx, column); err != nil {
				return ports.OperationReceipt{}, nil, err
			}
			payload, _ := json.Marshal(struct{ BoardID, ColumnID string }{
				input.BoardID.String(), id.String()})
			return ports.OperationReceipt{AggregateID: id, RowVersion: 1}, payload, nil
		})
}

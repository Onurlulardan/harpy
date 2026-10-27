package boards

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

type CreateBoardInput struct {
	Name, Description, InitialColumn string
}

func (service *Service) CreateBoard(
	ctx context.Context, scope ports.Scope, meta ports.Mutation, input CreateBoardInput,
) (ports.OperationReceipt, error) {
	name, columnName := strings.TrimSpace(input.Name), strings.TrimSpace(input.InitialColumn)
	if name == "" || columnName == "" || len(name) > 200 || len(columnName) > 200 ||
		len(input.Description) > 16_000 || meta.ExpectedVersion != 0 {
		return ports.OperationReceipt{}, ports.ErrInvalid
	}
	hash := ports.Fingerprint(name, input.Description, columnName)
	return service.mutator.Mutate(ctx, scope, "work.board.create", domain.ID{}, meta, hash,
		func(ctx context.Context, stores ports.Stores) (ports.OperationReceipt, []byte, error) {
			now := service.clock.Now().UTC()
			boardID, err := service.ids.New(now)
			if err != nil {
				return ports.OperationReceipt{}, nil, ports.ErrUnavailable
			}
			columnID, err := service.ids.New(now)
			if err != nil {
				return ports.OperationReceipt{}, nil, ports.ErrUnavailable
			}
			membershipID, err := service.ids.New(now)
			if err != nil {
				return ports.OperationReceipt{}, nil, ports.ErrUnavailable
			}
			board := domain.Board{ID: boardID, CreatorID: scope.ActorID(), Name: name,
				Description: input.Description, Lifecycle: domain.LifecycleActive,
				RowVersion: 1, CreatedAt: now}
			column := domain.Column{ID: columnID, BoardID: boardID, Name: columnName,
				Color: domain.DefaultColor, Status: domain.StatusOpen,
				Position: domain.SparseGap, RowVersion: 1}
			membership := domain.Membership{ID: membershipID, BoardID: boardID,
				Principal: domain.Principal{Kind: domain.PrincipalUser, ID: scope.ActorID()},
				Role:      domain.RoleAdmin, Active: true, RowVersion: 1}
			if !board.Valid() || !column.Valid() || !membership.Valid() {
				return ports.OperationReceipt{}, nil, ports.ErrInvalid
			}
			if err = stores.Boards().InsertBoard(ctx, board); err != nil {
				return ports.OperationReceipt{}, nil, err
			}
			if err = stores.Boards().InsertColumn(ctx, column); err != nil {
				return ports.OperationReceipt{}, nil, err
			}
			if err = stores.Boards().UpsertMembership(ctx, membership); err != nil {
				return ports.OperationReceipt{}, nil, err
			}
			payload, _ := json.Marshal(struct{ BoardID, Name string }{boardID.String(), name})
			return ports.OperationReceipt{AggregateID: boardID, RowVersion: 1}, payload, nil
		})
}

package columns

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

type UpdateInput struct {
	ColumnID domain.ID
	Name     string
	Color    string
	Status   domain.TaskStatus
}

func (service *Service) UpdateColumn(ctx context.Context, scope ports.Scope,
	meta ports.Mutation, input UpdateInput) (ports.OperationReceipt, error) {
	input.Name = strings.TrimSpace(input.Name)
	if !input.ColumnID.Valid() || input.Name == "" || len(input.Name) > 200 ||
		!domain.ValidColor(input.Color) || !input.Status.Valid() || meta.ExpectedVersion < 1 {
		return ports.OperationReceipt{}, ports.ErrInvalid
	}
	hash := ports.Fingerprint(input.ColumnID.String(), input.Name, input.Color, string(input.Status),
		ports.VersionString(meta.ExpectedVersion))
	return service.mutator.Mutate(ctx, scope, "work.column.update", input.ColumnID, meta, hash,
		func(ctx context.Context, stores ports.Stores) (ports.OperationReceipt, []byte, error) {
			column, err := stores.Boards().LoadColumnForUpdate(ctx, input.ColumnID)
			if err != nil {
				return ports.OperationReceipt{}, nil, err
			}
			if column.RowVersion != meta.ExpectedVersion {
				return ports.OperationReceipt{}, nil, ports.ErrConflict
			}
			column.Name, column.Color, column.Status = input.Name, input.Color, input.Status
			column.RowVersion++
			if !column.Valid() {
				return ports.OperationReceipt{}, nil, ports.ErrInvalid
			}
			if err = stores.Boards().PersistColumn(ctx, column, meta.ExpectedVersion); err != nil {
				return ports.OperationReceipt{}, nil, err
			}
			payload, _ := json.Marshal(struct {
				ColumnID string
				Version  int64
			}{column.ID.String(), column.RowVersion})
			return ports.OperationReceipt{AggregateID: column.ID, RowVersion: column.RowVersion}, payload, nil
		})
}

package columns

import (
	"context"
	"encoding/json"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

type DeleteInput struct {
	BoardID, ColumnID domain.ID
}

func (service *Service) DeleteColumn(ctx context.Context, scope ports.Scope,
	meta ports.Mutation, input DeleteInput,
) (ports.OperationReceipt, error) {
	if !input.BoardID.Valid() || !input.ColumnID.Valid() || meta.ExpectedVersion < 1 {
		return ports.OperationReceipt{}, ports.ErrInvalid
	}
	hash := ports.Fingerprint(input.BoardID.String(), input.ColumnID.String(),
		ports.VersionString(meta.ExpectedVersion))
	return service.mutator.Mutate(ctx, scope, "work.column.update", input.ColumnID,
		meta, hash, func(ctx context.Context, stores ports.Stores) (
			ports.OperationReceipt, []byte, error,
		) {
			column, err := stores.Boards().LoadColumnForUpdate(ctx, input.ColumnID)
			if err != nil || column.BoardID != input.BoardID ||
				column.RowVersion != meta.ExpectedVersion {
				return ports.OperationReceipt{}, nil, ports.ErrConflict
			}
			columns, err := stores.Boards().LoadColumnsForUpdate(ctx, input.BoardID)
			if err != nil || len(columns) <= 1 {
				return ports.OperationReceipt{}, nil, ports.ErrConflict
			}
			if err = stores.Boards().DeleteColumn(ctx, column.ID, meta.ExpectedVersion); err != nil {
				return ports.OperationReceipt{}, nil, err
			}
			version := meta.ExpectedVersion + 1
			payload, _ := json.Marshal(struct {
				BoardID, ColumnID string
				Version           int64
			}{input.BoardID.String(), input.ColumnID.String(), version})
			return ports.OperationReceipt{AggregateID: input.ColumnID,
				RowVersion: version}, payload, nil
		})
}

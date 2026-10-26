package postgres

import (
	"context"
	_ "embed"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

//go:embed sql/tenant/complete_operation.sql
var completeOperationSQL string

func (store *Store) PersistResult(ctx context.Context, operationID domain.ID, receipt ports.OperationReceipt, response []byte) error {
	if !operationID.Valid() || !receipt.AggregateID.Valid() || receipt.RowVersion < 1 || len(response) > 65536 {
		return application.ErrInvalid
	}
	return executeOne(ctx, store, completeOperationSQL, operationID, receipt.AggregateID,
		receipt.RowVersion, append([]byte{}, response...))
}

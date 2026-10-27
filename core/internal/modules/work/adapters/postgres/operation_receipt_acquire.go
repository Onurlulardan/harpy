package postgres

import (
	"bytes"
	"context"
	_ "embed"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
	"github.com/google/uuid"
)

//go:embed sql/tenant/claim_operation.sql
var claimOperationSQL string

//go:embed sql/tenant/load_operation_receipt.sql
var loadOperationReceiptSQL string

func (store *Store) Acquire(ctx context.Context, operationID domain.ID, key, kind string, hash [32]byte) (ports.OperationReceipt, bool, error) {
	if !operationID.Valid() || key == "" || kind == "" || len(key) > 200 {
		return ports.OperationReceipt{}, false, application.ErrInvalid
	}
	if err := store.check(ctx); err != nil {
		return ports.OperationReceipt{}, false, err
	}
	result, err := store.database.run(ctx, claimOperationSQL, operationID, kind,
		store.bridge.ActorID(), key, hash[:])
	if err != nil {
		return ports.OperationReceipt{}, false, err
	}
	if result.RowsAffected() == 1 {
		return ports.OperationReceipt{}, false, nil
	}
	var storedHash []byte
	var state string
	var aggregateID uuid.NullUUID
	var version int64
	err = store.database.QueryRow(ctx, loadOperationReceiptSQL, store.bridge.ActorID(), kind, key).
		Scan(&storedHash, &state, &aggregateID, &version)
	if err != nil || !bytes.Equal(storedHash, hash[:]) || state != "COMPLETED" || !aggregateID.Valid || version < 1 {
		return ports.OperationReceipt{}, false, application.ErrConflict
	}
	return ports.OperationReceipt{AggregateID: domain.ID(aggregateID.UUID), RowVersion: version}, true, nil
}

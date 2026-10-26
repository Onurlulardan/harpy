package profilebackfill

import (
	"bytes"
	"context"
	"crypto/sha256"

	profilebackfillapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profilebackfill"
)

func (cursor *recoveryInventoryCursor) Fetch(
	ctx context.Context,
) ([]RecoveryInventoryRow, error) {
	if cursor == nil || !cursor.usable(scopeUsable(ctx, cursor.scope)) {
		return nil, profilebackfillapp.ErrInvalid
	}
	rows, err := cursor.store.fetchRecoveryInventory(ctx)
	if err != nil {
		return nil, recoveryStoreError(err)
	}
	defer rows.Close()
	return scanRecoveryInventoryRows(rows)
}

func scanRecoveryInventoryRows(
	rows recoveryInventoryRows,
) ([]RecoveryInventoryRow, error) {
	result := make([]RecoveryInventoryRow, 0,
		profilebackfillapp.RecoveryInventoryReadBatchSize)
	for rows.Next() {
		row, err := scanRecoveryInventoryRow(rows)
		if err != nil || len(result) == cap(result) ||
			(len(result) > 0 && bytes.Compare(
				result[len(result)-1].Resource.ID[:], row.Resource.ID[:],
			) >= 0) {
			clear(row.Resource.Document)
			wipeRecoveryInventoryResult(result)
			return nil, profilebackfillapp.ErrUnavailable
		}
		result = append(result, row)
	}
	if rows.Err() != nil {
		wipeRecoveryInventoryResult(result)
		return nil, profilebackfillapp.ErrUnavailable
	}
	return result, nil
}

func scanRecoveryInventoryRow(
	rows recoveryInventoryRows,
) (RecoveryInventoryRow, error) {
	var row RecoveryInventoryRow
	var digest []byte
	err := rows.Scan(
		&row.Resource.ID, &row.Resource.SchemaVersionID, &row.SchemaVersion,
		&row.SchemaState, &row.LineageKind, &row.Resource.Revision,
		&row.Resource.Document, &digest,
	)
	if err != nil || len(digest) != sha256.Size {
		clear(row.Resource.Document)
		clear(digest)
		return RecoveryInventoryRow{}, profilebackfillapp.ErrUnavailable
	}
	copy(row.Resource.SHA256[:], digest)
	clear(digest)
	if row.Validate() != nil {
		clear(row.Resource.Document)
		return RecoveryInventoryRow{}, profilebackfillapp.ErrUnavailable
	}
	return row, nil
}

type recoveryInventoryRows interface {
	Next() bool
	Scan(...any) error
	Err() error
}

func wipeRecoveryInventoryResult(rows []RecoveryInventoryRow) {
	for index := range rows {
		clear(rows[index].Resource.Document)
		rows[index].Resource.Document = nil
	}
}

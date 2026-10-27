package systemtransfer

import (
	"context"

	app "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
)

func (storage *store) ListReceipts(ctx context.Context, read app.ControlRead,
	after int64, limit int) (app.ReceiptPage, error) {
	if !storage.usable(ctx) || read.Validate() != nil || after < 0 ||
		after > 1_000_000 || limit < 1 || limit > 100 ||
		read.Evidence.TenantID != storage.tenantID ||
		read.Evidence.ActorID != storage.actorID {
		return app.ReceiptPage{}, app.ErrControlUnavailable
	}
	result, err := storage.query(ctx, readReceiptsSQL, read.TransferID,
		storage.actorID, string(read.Evidence.Resource), read.Evidence.TokenVersion,
		read.Evidence.AuthorizationVersion, after, limit+1)
	if err != nil {
		return app.ReceiptPage{}, mapError(err, storage.notFound)
	}
	defer result.Close()
	page := app.ReceiptPage{Rows: make([]app.ReceiptRecord, 0, limit)}
	for result.Next() {
		var receipt app.ReceiptRecord
		if err = result.Scan(&receipt.RowIndex, &receipt.Outcome,
			&receipt.OwnerKind, &receipt.OwnerID, &receipt.ErrorCode,
			&receipt.SourceSHA256, &receipt.ResultSHA256); err != nil {
			return app.ReceiptPage{}, mapError(err, storage.notFound)
		}
		if len(page.Rows) == limit {
			page.HasMore = true
			break
		}
		page.Rows = append(page.Rows, receipt)
		page.NextAfter = receipt.RowIndex
	}
	if err = result.Err(); err != nil {
		return app.ReceiptPage{}, mapError(err, storage.notFound)
	}
	return page, nil
}

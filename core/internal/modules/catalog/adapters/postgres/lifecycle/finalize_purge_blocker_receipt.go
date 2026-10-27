package lifecycle

import (
	"context"
	_ "embed"

	catalogapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/lifecycle"
)

//go:embed sql/tenant/record_purge_blocker_receipt.sql
var recordPurgeBlockerReceiptSQL string

func (storage *store) FinalizePurgeBlockerReceipt(
	ctx context.Context, receipt catalogapp.PurgeBlockerReceipt,
) error {
	if !validStableID(receipt.ReceiptID) || !validStableID(receipt.OperationID) ||
		!validStableID(receipt.ComponentID) || receipt.ReceiptID == receipt.OperationID ||
		receipt.ReferenceEpoch < 1 || receipt.CatalogVersion == "" ||
		!digestPattern.MatchString(receipt.ZeroDigest) || !canonicalTime(receipt.CompletedAt) {
		return catalogapp.ErrInvalid
	}
	if !storage.usable(ctx, true) {
		return unavailable()
	}
	var stored catalogapp.PurgeBlockerReceipt
	err := storage.queryRow(ctx, recordPurgeBlockerReceiptSQL, receipt.ReceiptID,
		receipt.OperationID, receipt.ComponentID, receipt.ReferenceEpoch,
		receipt.CatalogVersion, receipt.ZeroDigest, receipt.CompletedAt).Scan(
		&stored.ReceiptID, &stored.OperationID, &stored.ComponentID, &stored.ReferenceEpoch,
		&stored.CatalogVersion, &stored.ZeroDigest, &stored.CompletedAt)
	if err != nil || stored != receipt {
		return unavailable()
	}
	return nil
}

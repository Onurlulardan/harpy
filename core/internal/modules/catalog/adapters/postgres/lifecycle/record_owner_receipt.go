package lifecycle

import (
	"context"
	_ "embed"

	catalogapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/lifecycle"
	catalogdomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/lifecycle"
)

//go:embed sql/tenant/record_owner_receipt.sql
var recordOwnerReceiptSQL string

func (storage *store) RecordOwnerReceipt(
	ctx context.Context, receipt catalogdomain.OwnerReceipt,
) error {
	if !validOwnerReceipt(receipt) {
		return catalogapp.ErrInvalid
	}
	if !storage.usable(ctx, true) {
		return unavailable()
	}
	stored, err := scanOwnerReceipt(storage.queryRow(ctx, recordOwnerReceiptSQL,
		receipt.ReceiptID.String(), receipt.OperationID.String(), receipt.ComponentID.String(),
		receipt.Owner, receipt.OwnerGeneration, receipt.Contract, receipt.ContentSHA256,
		receipt.Action, receipt.Ready, receipt.CompletedAt, receiptDigest(receipt)))
	if err != nil || stored != receipt {
		return unavailable()
	}
	return nil
}

func validOwnerReceipt(value catalogdomain.OwnerReceipt) bool {
	if !value.ReceiptID.Valid() || !value.OperationID.Valid() || !value.ComponentID.Valid() ||
		value.ReceiptID == value.OperationID || value.Owner == "" || value.OwnerGeneration < 1 ||
		value.Contract == "" || !digestPattern.MatchString(value.ContentSHA256) || !value.Ready ||
		!canonicalTime(value.CompletedAt) {
		return false
	}
	switch value.Action {
	case "stage", "cleanup", "activate", "disable", "enable", "retire", "restore", "purge":
		return true
	default:
		return false
	}
}

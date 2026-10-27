package lifecycle

import (
	"context"
	_ "embed"

	catalogapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/lifecycle"
	catalogdomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/lifecycle"
	"github.com/google/uuid"
)

//go:embed sql/tenant/record_command_receipt.sql
var recordCommandReceiptSQL string

func (storage *store) applyCommandReceipts(
	ctx context.Context,
	mutation catalogapp.InstallationMutation,
	installation catalogdomain.Installation,
) error {
	bindingState, ok := bindingStateFor(installation.State)
	if !ok || len(mutation.OwnerReceipts) > 256 {
		return catalogapp.ErrInvalid
	}
	installationID, _ := uuid.Parse(installation.ID.String())
	for _, receipt := range mutation.OwnerReceipts {
		if !validOwnerReceipt(receipt) || receipt.Action != "disable" && receipt.Action != "enable" {
			return catalogapp.ErrInvalid
		}
		var returned string
		err := storage.queryRow(ctx, recordCommandReceiptSQL, receipt.ReceiptID.String(),
			receipt.OperationID.String(), installation.PackageID.String(), receipt.ComponentID.String(),
			receipt.Owner, receipt.OwnerGeneration, receipt.Contract, receipt.ContentSHA256,
			receipt.Action, mutation.IdempotencyKey, receipt.CompletedAt).Scan(&returned)
		if err != nil || returned != receipt.ReceiptID.String() {
			return unavailable()
		}
		if _, err = storage.execute(ctx, upsertComponentBindingSQL, installationID,
			receipt.ComponentID.String(), receipt.Owner, receipt.OwnerGeneration,
			receipt.Contract, receipt.ContentSHA256, bindingState); err != nil {
			return unavailable()
		}
	}
	return nil
}

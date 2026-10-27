package restoreexecution

import (
	"context"
	"crypto/sha256"
	"encoding/hex"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func (service *Service) ImportReceipt(
	ctx context.Context,
	record ports.RestoreReceiptImport,
) error {
	digest := sha256.Sum256(record.ReceiptBytes)
	if service == nil || service.store == nil || ctx == nil ||
		ctx.Err() != nil ||
		!canonicalUUID.MatchString(record.OperationID) ||
		!canonicalUUID.MatchString(record.RecoverySetID) ||
		!canonicalUUID.MatchString(record.EnvironmentID) ||
		!canonicalUUID.MatchString(record.NewDatabaseSystemID) ||
		!digestPattern.MatchString(record.PlanSHA256) ||
		!digestPattern.MatchString(record.ManifestSHA256) ||
		record.ReceiptSHA256 != hex.EncodeToString(digest[:]) ||
		!digestPattern.MatchString(record.ObjectMerkleSHA256) ||
		!digestPattern.MatchString(record.LocatorRemapSHA256) ||
		!digestPattern.MatchString(record.SafetyEvidenceSHA256) ||
		!digestPattern.MatchString(record.DrainAckSHA256) ||
		!digestPattern.MatchString(record.SwitchAckSHA256) ||
		!digestPattern.MatchString(record.ControlDatabaseLocatorSHA256) ||
		!digestPattern.MatchString(record.ScratchDatabaseLocatorSHA256) ||
		record.ControlDatabaseLocatorSHA256 ==
			record.ScratchDatabaseLocatorSHA256 ||
		!validComponentInventory(record.ComponentInventory) {
		return ErrTransition
	}
	value, err := service.store.ImportRestoreReceipt(ctx, record)
	return changed(err, value)
}

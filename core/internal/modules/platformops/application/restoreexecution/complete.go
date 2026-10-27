package restoreexecution

import (
	"context"
	"crypto/sha256"
	"encoding/hex"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func (service *Service) Complete(
	ctx context.Context,
	record ports.RestoreCompletion,
) error {
	digest := sha256.Sum256(record.ReceiptBytes)
	if service == nil || service.store == nil || ctx == nil ||
		ctx.Err() != nil || !record.RestoreExecutionIdentity.Valid() ||
		record.ReceiptSHA256 != hex.EncodeToString(digest[:]) {
		return ErrTransition
	}
	value, err := service.store.CompleteRestoreExecution(ctx, record)
	return changed(err, value)
}

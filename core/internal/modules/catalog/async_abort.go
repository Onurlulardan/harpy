package catalog

import (
	"context"
	"crypto/sha256"
	"encoding/hex"

	catalogapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/lifecycle"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (port *AsyncPort) Abort(
	ctx context.Context, scope platformtx.TenantScope, operationID uuid.UUID,
) error {
	if operationID == uuid.Nil || operationID.Version() != 7 {
		return catalogapp.ErrInvalid
	}
	writer, _, err := port.bindAsync(ctx, scope)
	if err != nil {
		return err
	}
	now := port.clock.Now()
	digest := sha256.Sum256([]byte("catalog.operation.cancel.v1\x00" + operationID.String()))
	request, err := asyncpostgres.NewOwnedJobCancellation(operationID, now,
		"CATALOG_OPERATION_CANCELLED", hex.EncodeToString(digest[:]))
	if err != nil {
		return asyncUnavailable()
	}
	if _, err = writer.CancelOwnedJob(ctx, request); err != nil {
		return asyncUnavailable()
	}
	return nil
}

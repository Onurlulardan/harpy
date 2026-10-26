package demohandoff

import (
	"context"
	"time"

	api "github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/demohandoff"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

var zeroSHA256 [32]byte

func requestValid(request api.Request) bool {
	return stableV7(uuid.UUID(request.ProvisionReceiptID)) &&
		stableV7(uuid.UUID(request.ComponentID)) &&
		uuid.UUID(request.PackageID) != uuid.Nil &&
		stableV7(uuid.UUID(request.ReleaseID)) &&
		stableV7(uuid.UUID(request.FormsRunID)) &&
		request.RequestSHA256 != zeroSHA256 &&
		request.FormCount > 0 && request.RowCount > 0 &&
		request.RowCount >= request.FormCount &&
		request.ReadbackSHA256 != zeroSHA256 &&
		canonicalTime(request.CompletedAt)
}

func receiptMatches(receipt api.Receipt, request api.Request) bool {
	return stableV7(uuid.UUID(receipt.HandoffID)) &&
		receipt.ProvisionReceiptID == request.ProvisionReceiptID &&
		receipt.ComponentID == request.ComponentID &&
		receipt.PackageID == request.PackageID &&
		receipt.ReleaseID == request.ReleaseID &&
		receipt.FormsRunID == request.FormsRunID &&
		receipt.RequestSHA256 == request.RequestSHA256 &&
		receipt.FormCount == request.FormCount &&
		receipt.RowCount == request.RowCount &&
		receipt.ReadbackSHA256 == request.ReadbackSHA256 &&
		receipt.CompletedAt.Equal(request.CompletedAt)
}

func scopeValid(ctx context.Context, scope platformtx.TenantScope) bool {
	return ctx != nil && ctx.Err() == nil && !platformtx.NetworkAllowed(ctx) &&
		platformtx.ContextOwnsScope(ctx, scope) &&
		scope.Mode() == platformtx.Command &&
		scope.OperationID() == api.Operation &&
		platformtx.RequireOwner(scope, owner) == nil
}

func stableV7(value uuid.UUID) bool {
	return value != uuid.Nil && value.Version() == 7 &&
		value.Variant() == uuid.RFC4122
}

func canonicalTime(value time.Time) bool {
	return !value.IsZero() && value.Year() >= 2020 &&
		value.Location() == time.UTC && value.Nanosecond()%1_000 == 0
}

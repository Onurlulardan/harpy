package identity

import (
	"context"

	identitydomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/identity"
)

type BulkDeleteReceiptStore[Scope any] interface {
	ReserveBulkDelete(
		context.Context, Scope, identitydomain.BulkDeleteReceiptCommand,
	) (bool, error)
}

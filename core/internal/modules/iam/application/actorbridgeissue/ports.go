package actorbridgeissue

import (
	"context"

	actorassertionv1 "github.com/Onurlulardan/harpy/contracts/actorassertion/v1"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type TenantScope = platformtx.TenantScope

type Store interface {
	Resolve(context.Context, TenantScope, actorassertionv1.Domain) (Evidence, error)
}

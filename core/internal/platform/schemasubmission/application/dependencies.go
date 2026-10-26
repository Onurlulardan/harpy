package application

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type Store interface {
	Submit(context.Context, PreparedSubmission) (StoredReceipt, error)
}

type StoreBinder interface {
	Bind(context.Context, platformtx.TenantScope) (Store, error)
}

type StoredReceipt struct {
	JobID      string
	Generation int64
	PlanSHA256 string
	Replay     bool
}

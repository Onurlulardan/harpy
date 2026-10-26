package execution

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type Row interface {
	Scan(destinations ...any) error
}

type Rows interface {
	Close()
	Err() error
	Next() bool
	Scan(destinations ...any) error
}

type QueryDriver interface {
	Owner() string
	Execute(ctx context.Context, statement string, arguments ...any) error
	Query(ctx context.Context, statement string, arguments ...any) (Rows, error)
	QueryRow(ctx context.Context, statement string, arguments ...any) Row
}

// SnapshotDriver is implemented mechanically by a tenant owner PostgreSQL adapter.
type SnapshotDriver interface {
	QueryDriver
	Scope() transaction.TenantScope
}

// GlobalSnapshotDriver is the distinct global Billing read-only driver ABI.
type GlobalSnapshotDriver interface {
	QueryDriver
	GlobalScope() transaction.GlobalScope
}

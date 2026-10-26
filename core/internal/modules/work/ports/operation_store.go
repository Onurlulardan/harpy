package ports

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
)

type OperationReceipt struct {
	AggregateID domain.ID
	RowVersion  int64
}

type OperationStore interface {
	Acquire(context.Context, domain.ID, string, string, [32]byte) (OperationReceipt, bool, error)
	PersistResult(context.Context, domain.ID, OperationReceipt, []byte) error
	AppendAudit(context.Context, domain.ID, domain.ID, string, domain.ID, domain.ID, int64) (int64, error)
}

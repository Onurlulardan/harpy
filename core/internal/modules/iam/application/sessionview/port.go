package sessionview

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type Port interface {
	Read(
		context.Context, platformtx.TenantScope, uuid.UUID, uuid.UUID,
	) (View, error)
}

type Reader struct{ port Port }

func NewReader(port Port) (*Reader, error) {
	if port == nil {
		return nil, ErrContract
	}
	return &Reader{port: port}, nil
}

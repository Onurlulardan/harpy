package installationcredentialpostgres

import (
	"context"

	installationcredential "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/installationcredential"
)

type Row interface{ Scan(...any) error }

type Executor struct {
	QueryRow func(context.Context, string, ...any) Row
}

type store struct {
	queryRow func(context.Context, string, ...any) Row
}

func NewStore(executor Executor) (installationcredential.Store, error) {
	if executor.QueryRow == nil {
		return nil, installationcredential.ErrStorage
	}
	return &store{queryRow: executor.QueryRow}, nil
}

var _ installationcredential.Store = (*store)(nil)

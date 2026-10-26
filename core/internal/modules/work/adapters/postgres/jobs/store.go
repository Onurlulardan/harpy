package jobs

import (
	"context"

	jobapp "github.com/Onurlulardan/harpy/core/internal/modules/work/application/jobs"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

type row interface{ Scan(...any) error }

type rows interface {
	Next() bool
	Scan(...any) error
	Err() error
	Close()
}

type databaseAccess struct {
	execute  func(context.Context, string, ...any) (int64, error)
	query    func(context.Context, string, ...any) (rows, error)
	queryRow func(context.Context, string, ...any) row
}

type Store struct{ database databaseAccess }

func (store *Store) valid() bool {
	return store != nil && store.database.execute != nil && store.database.query != nil &&
		store.database.queryRow != nil
}

func one(affected int64, err error) error {
	if err != nil {
		return ports.ErrUnavailable
	}
	if affected != 1 {
		return ports.ErrConflict
	}
	return nil
}

var _ jobapp.Store = (*Store)(nil)

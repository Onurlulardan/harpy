package customerworker

import (
	"context"

	workerapp "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/customerworker"
)

type Row interface{ Scan(...any) error }
type QueryRow func(context.Context, string, ...any) Row
type RowNotFound func(error) bool

type Store struct {
	queryRow QueryRow
	notFound RowNotFound
}

func NewStore(queryRow QueryRow, notFound RowNotFound) (*Store, error) {
	if queryRow == nil || notFound == nil {
		return nil, workerapp.ErrUnavailable
	}
	return &Store{queryRow: queryRow, notFound: notFound}, nil
}

var _ workerapp.ClaimStore = (*Store)(nil)

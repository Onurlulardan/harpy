package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/audit/application"
)

type row interface{ Scan(...any) error }

type rows interface {
	Next() bool
	Scan(...any) error
	Err() error
	Close()
}

type Store struct {
	scopeAccess application.ScopeAccess
	usable      func(context.Context) bool
	queryRow    func(context.Context, string, ...any) row
	query       func(context.Context, string, ...any) (rows, error)
	notFound    func(error) bool
}

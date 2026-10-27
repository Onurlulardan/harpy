package compatibility

import (
	"context"

	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
)

type row = platformpostgres.Row

type Database interface {
	QueryRow(context.Context, string, ...any) platformpostgres.Row
}

type database = Database

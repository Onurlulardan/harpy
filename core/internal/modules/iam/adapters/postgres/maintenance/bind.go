package maintenance

import (
	"context"

	maintenanceapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/maintenance"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
)

type DatabaseAccess struct {
	Query    func(context.Context, string, ...any) (platformpostgres.Rows, error)
	QueryRow func(context.Context, string, ...any) platformpostgres.Row
}

func (access DatabaseAccess) valid() bool {
	return access.Query != nil && access.QueryRow != nil
}

type databaseAccess struct{ DatabaseAccess }

type store struct {
	database databaseAccess
	runtime  *Runtime
}

func NewStore(
	runtime *Runtime,
	access DatabaseAccess,
) maintenanceapp.Store {
	if runtime == nil || !runtime.usable() || !access.valid() {
		return nil
	}
	return &store{database: databaseAccess{access}, runtime: runtime}
}

func (database databaseAccess) Query(
	ctx context.Context, query string, arguments ...any,
) (platformpostgres.Rows, error) {
	return database.DatabaseAccess.Query(ctx, query, arguments...)
}

func (database databaseAccess) QueryRow(
	ctx context.Context, query string, arguments ...any,
) platformpostgres.Row {
	return database.DatabaseAccess.QueryRow(ctx, query, arguments...)
}

var _ maintenanceapp.Store = (*store)(nil)

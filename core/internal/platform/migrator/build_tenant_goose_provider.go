package migrator

import (
	"database/sql"
	"fmt"
	"path"

	postgresstore "github.com/Onurlulardan/harpy/core/internal/platform/migrator/adapters/postgres"
	"github.com/pressly/goose/v3"
)

func (runner *Runner) buildTenantGooseProvider(
	database *sql.DB,
	target postgresstore.TenantMigrationTarget,
	session postgresstore.TenantSessionProof,
	runID string,
) (*goose.Provider, error) {
	store, err := postgresstore.NewTenantStore(
		runID, target, session, runner.tenantSelected,
	)
	if err != nil {
		return nil, err
	}
	provider, err := goose.NewProvider(
		goose.DialectCustom, database, runner.tenantFiles,
		goose.WithStore(store), goose.WithAllowOutofOrder(false),
		goose.WithDisableGlobalRegistry(true), goose.WithIsolateDDL(false),
		goose.WithSlog(runner.logger),
	)
	if err != nil {
		return nil, fmt.Errorf("create Goose tenant provider: %w", err)
	}
	sources := provider.ListSources()
	if len(sources) != len(runner.tenantSelected) {
		return nil, fmt.Errorf("Goose tenant source count differs from verified selection")
	}
	for index, source := range sources {
		selected := runner.tenantSelected[index]
		if source.Type != goose.TypeSQL || source.Version != selected.Version ||
			path.Base(source.Path) != path.Base(selected.Path) {
			return nil, fmt.Errorf("Goose tenant source %d differs from selection", index)
		}
	}
	return provider, nil
}

package migrator

import (
	"fmt"
	"path"

	postgresstore "github.com/Onurlulardan/harpy/core/internal/platform/migrator/adapters/postgres"
	"github.com/pressly/goose/v3"
)

func (runner *Runner) buildGooseProvider(runID string) (*goose.Provider, error) {
	store, err := postgresstore.NewStore(runID, runner.selected)
	if err != nil {
		return nil, err
	}
	provider, err := goose.NewProvider(
		goose.DialectCustom,
		runner.database,
		runner.platformFiles,
		goose.WithStore(store),
		goose.WithAllowOutofOrder(false),
		goose.WithDisableGlobalRegistry(true),
		goose.WithIsolateDDL(false),
		goose.WithSlog(runner.logger),
	)
	if err != nil {
		return nil, fmt.Errorf("create Goose platform provider: %w", err)
	}
	sources := provider.ListSources()
	if len(sources) != len(runner.selected) {
		return nil, fmt.Errorf("Goose source count differs from verified selection")
	}
	for index, source := range sources {
		selected := runner.selected[index]
		if source.Type != goose.TypeSQL || source.Version != selected.Version ||
			path.Base(source.Path) != path.Base(selected.Path) {
			return nil, fmt.Errorf("Goose source %d differs from verified selection", index)
		}
	}
	return provider, nil
}

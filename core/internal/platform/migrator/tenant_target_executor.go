package migrator

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"log/slog"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/migration"
	postgresstore "github.com/Onurlulardan/harpy/core/internal/platform/migrator/adapters/postgres"
)

type TenantTargetExecutor struct{ runner *Runner }

type TenantTargetDependencies struct {
	Database             func() (*sql.DB, error)
	Files                fs.FS
	Selected             []postgresv1.SelectedMigration
	HistorySupersessions []migration.HistorySupersession
	Logger               *slog.Logger
}

func NewTenantTargetExecutor(
	dependencies TenantTargetDependencies,
) (*TenantTargetExecutor, error) {
	if dependencies.Database == nil || dependencies.Files == nil ||
		dependencies.Logger == nil {
		return nil, fmt.Errorf("tenant target migration dependencies are incomplete")
	}
	if err := validateTenantSelection(dependencies.Selected); err != nil {
		return nil, err
	}
	aliases, err := indexHistorySupersessions(
		dependencies.Selected, dependencies.HistorySupersessions,
	)
	if err != nil {
		return nil, err
	}
	runner := &Runner{tenantDatabase: dependencies.Database,
		tenantFiles: dependencies.Files,
		tenantSelected: append([]postgresv1.SelectedMigration(nil),
			dependencies.Selected...),
		tenantAliases: aliases,
		logger:        dependencies.Logger}
	return &TenantTargetExecutor{runner: runner}, nil
}

func (executor *TenantTargetExecutor) Execute(
	ctx context.Context, tenantID, schemaName, runID string,
) ([]int64, error) {
	if executor == nil || executor.runner == nil || tenantID == "" ||
		schemaName == "" || runID == "" {
		return nil, fmt.Errorf("tenant target migration request is invalid")
	}
	return executor.runner.executeTenantTarget(ctx,
		postgresstore.TenantMigrationTarget{TenantID: tenantID,
			SchemaName: schemaName}, runID)
}

package migrator

import (
	"context"
	"database/sql"
	"fmt"

	postgresstore "github.com/Onurlulardan/harpy/core/internal/platform/migrator/adapters/postgres"
)

type historyRow struct {
	version                      int64
	recordKind                   string
	isApplied                    bool
	owner                        string
	scope                        string
	origin                       string
	selectedPath                 string
	selectedSHA256               string
	runID                        string
	targetDatabaseGeneration     int64
	nextMinimumRuntimeGeneration int64
}

func readHistory(
	ctx context.Context,
	connection *sql.Conn,
) (map[int64]historyRow, bool, error) {
	rows, exists, err := postgresstore.ReadAppliedHistory(ctx, connection)
	if err != nil {
		return nil, exists, err
	}
	if !exists {
		return map[int64]historyRow{}, false, nil
	}
	history := make(map[int64]historyRow)
	for _, applied := range rows {
		row := historyRow{
			version: applied.Version, recordKind: applied.RecordKind,
			isApplied: applied.IsApplied, owner: applied.Owner, scope: applied.Scope,
			origin: applied.Origin, selectedPath: applied.SelectedPath,
			selectedSHA256: applied.SelectedSHA256, runID: applied.RunID,
			targetDatabaseGeneration:     applied.TargetDatabaseGeneration,
			nextMinimumRuntimeGeneration: applied.NextMinimumRuntimeGeneration,
		}
		if _, duplicate := history[row.version]; duplicate {
			return nil, true, fmt.Errorf("migration history version is duplicated")
		}
		history[row.version] = row
	}
	return history, true, nil
}

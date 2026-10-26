package postgres

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
)

//go:embed sql/history_table_exists.sql
var historyTableExistsSQL string

//go:embed sql/read_applied_history.sql
var readAppliedHistorySQL string

type AppliedMigration struct {
	Version                      int64
	RecordKind                   string
	IsApplied                    bool
	Owner                        string
	Scope                        string
	Origin                       string
	SelectedPath                 string
	SelectedSHA256               string
	RunID                        string
	TargetDatabaseGeneration     int64
	NextMinimumRuntimeGeneration int64
}

func ReadAppliedHistory(
	ctx context.Context,
	connection *sql.Conn,
) ([]AppliedMigration, bool, error) {
	var exists bool
	if err := connection.QueryRowContext(ctx, historyTableExistsSQL).Scan(&exists); err != nil {
		return nil, false, fmt.Errorf("inspect migration history table: %w", err)
	}
	if !exists {
		return nil, false, nil
	}
	rows, err := connection.QueryContext(ctx, readAppliedHistorySQL)
	if err != nil {
		return nil, true, fmt.Errorf("read migration history: %w", err)
	}
	defer rows.Close()
	history := make([]AppliedMigration, 0)
	for rows.Next() {
		var row AppliedMigration
		if err := rows.Scan(
			&row.Version, &row.RecordKind, &row.IsApplied, &row.Owner, &row.Scope,
			&row.Origin, &row.SelectedPath, &row.SelectedSHA256, &row.RunID,
			&row.TargetDatabaseGeneration, &row.NextMinimumRuntimeGeneration,
		); err != nil {
			return nil, true, fmt.Errorf("scan migration history: %w", err)
		}
		history = append(history, row)
	}
	if err := rows.Err(); err != nil {
		return nil, true, fmt.Errorf("iterate migration history: %w", err)
	}
	return history, true, nil
}

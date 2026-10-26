package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/jackc/pgx/v5"
)

type actorBridgeTestQueryer struct {
	row       pgx.Row
	query     string
	arguments []any
}

func (queryer *actorBridgeTestQueryer) QueryRow(
	_ context.Context, query string, arguments ...any,
) pgx.Row {
	queryer.query = query
	queryer.arguments = append([]any(nil), arguments...)
	return queryer.row
}

type actorBridgeTestRow struct {
	scan func(...any) error
}

func (row actorBridgeTestRow) Scan(destinations ...any) error {
	return row.scan(destinations...)
}

func setNullTime(destination any, value time.Time, valid bool) {
	target := destination.(*sql.NullTime)
	target.Time = value
	target.Valid = valid
}

package postgres

import (
	"context"
	"fmt"
	"time"

	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/jackc/pgx/v5/pgtype"
)

type boundLifecycleDatabase struct {
	claim      JobClaim
	startedAt  time.Time
	completed  time.Time
	statements []string
}

func (database *boundLifecycleDatabase) Execute(
	context.Context,
	string,
	...any,
) (platformpostgres.CommandResult, error) {
	return platformpostgres.CommandResult{}, fmt.Errorf("unexpected bound execute")
}

func (database *boundLifecycleDatabase) QueryRow(
	_ context.Context,
	statement string,
	_ ...any,
) platformpostgres.Row {
	database.statements = append(database.statements, statement)
	return boundLifecycleRow{scan: func(destinations ...any) error {
		switch statement {
		case lockJobClaimSQL:
			if len(destinations) != 3 {
				return fmt.Errorf("lock destination count = %d", len(destinations))
			}
			*destinations[0].(*pgtype.Timestamptz) = pgtype.Timestamptz{}
			*destinations[1].(*pgtype.Timestamptz) = pgtype.Timestamptz{}
			*destinations[2].(*time.Time) = database.claim.leaseUntil
		case completeJobAttemptSQL, completeDatabaseJobSQL:
			*destinations[0].(*time.Time) = database.completed
		case markExternalJobEffectSQL, markExternalAttemptEffectSQL:
			*destinations[0].(*time.Time) = database.startedAt
		default:
			return fmt.Errorf("unexpected bound query")
		}
		return nil
	}}
}

type boundLifecycleRow struct {
	scan func(...any) error
}

func (row boundLifecycleRow) Scan(destinations ...any) error {
	return row.scan(destinations...)
}

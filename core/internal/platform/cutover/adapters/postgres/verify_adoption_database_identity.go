package postgres

import (
	"context"
	"fmt"

	cutoverv1 "github.com/Onurlulardan/harpy/contracts/cutover/v1"
	"github.com/jackc/pgx/v5"
)

func verifyAdoptionDatabaseIdentity(
	ctx context.Context,
	connections []*pgx.Conn,
	expected cutoverv1.DatabaseIdentity,
) error {
	for _, connection := range connections {
		observation, err := observeSession(ctx, connection)
		if err != nil || observation.validateDatabase(expected) != nil {
			return fmt.Errorf("cutover PostgreSQL database authority differs")
		}
	}
	return nil
}

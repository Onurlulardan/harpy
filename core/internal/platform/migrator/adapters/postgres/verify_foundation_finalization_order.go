package postgres

import (
	"context"
	_ "embed"
	"fmt"

	goosedb "github.com/pressly/goose/v3/database"
)

//go:embed sql/verify_foundation_finalization_order.sql
var verifyFoundationFinalizationOrderSQL string

func FoundationFinalizationOrderVerified(
	ctx context.Context,
	database goosedb.DBTxConn,
) (bool, error) {
	var verified bool
	if err := database.QueryRowContext(
		ctx, verifyFoundationFinalizationOrderSQL,
	).Scan(&verified); err != nil {
		return false, fmt.Errorf("verify foundation finalization order: %w", err)
	}
	return verified, nil
}

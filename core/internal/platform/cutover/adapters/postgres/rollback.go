package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

func rollbackBounded(tx pgx.Tx, timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	_ = tx.Rollback(ctx)
}

package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

func rollbackProbeTransaction(transaction pgx.Tx, timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	_ = transaction.Rollback(ctx)
}

func closeProbeConnection(connection *pgx.Conn, timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	_ = connection.Close(ctx)
}

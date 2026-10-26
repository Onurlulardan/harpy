package postgres

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/google/uuid"
)

func (pool *Pool) AdmitAnalyticsPublic(
	ctx context.Context, tenantID, releaseID uuid.UUID, generation int64,
	trustedClientIP string, occurredAt time.Time,
) (uuid.UUID, error) {
	if pool == nil || ctx == nil || ctx.Err() != nil || tenantID == uuid.Nil ||
		releaseID == uuid.Nil || generation < 1 || net.ParseIP(trustedClientIP) == nil ||
		occurredAt.IsZero() {
		return uuid.Nil, fmt.Errorf("analytics public admission input is invalid")
	}
	database, err := pool.BeginRuntimeTransaction(ctx, TransactionCommand)
	if err != nil {
		return uuid.Nil, err
	}
	var leaseID uuid.UUID
	err = database.transaction.QueryRow(ctx, admitAnalyticsPublicSQL, tenantID,
		releaseID, generation, trustedClientIP, occurredAt.UTC()).Scan(&leaseID)
	if err != nil || leaseID == uuid.Nil {
		_ = database.Rollback()
		if err != nil {
			return uuid.Nil, err
		}
		return uuid.Nil, fmt.Errorf("analytics public lease is invalid")
	}
	if err = database.Commit(ctx); err != nil {
		return uuid.Nil, err
	}
	return leaseID, nil
}

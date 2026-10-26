package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type AnalyticsPublicRoute struct {
	PublicID            [16]byte
	TenantID            uuid.UUID
	ReleaseID           uuid.UUID
	ResourceKind        string
	ResourceVersionID   uuid.UUID
	Audience            string
	State               string
	Generation          int64
	CapabilitySHA256    []byte
	CapabilityExpiresAt *time.Time
}

func (pool *Pool) ResolveAnalyticsPublicRoute(
	ctx context.Context, publicID [16]byte,
) (AnalyticsPublicRoute, error) {
	if pool == nil || ctx == nil || ctx.Err() != nil || publicID == [16]byte{} {
		return AnalyticsPublicRoute{}, fmt.Errorf("analytics public route input is invalid")
	}
	database, err := pool.BeginRuntimeTransaction(ctx, TransactionSnapshotQuery)
	if err != nil {
		return AnalyticsPublicRoute{}, err
	}
	var route AnalyticsPublicRoute
	var storedID []byte
	err = database.transaction.QueryRow(ctx, resolveAnalyticsPublicRouteSQL, publicID[:]).Scan(
		&storedID, &route.TenantID, &route.ReleaseID, &route.ResourceKind,
		&route.ResourceVersionID, &route.Audience, &route.State, &route.Generation,
		&route.CapabilitySHA256, &route.CapabilityExpiresAt,
	)
	if err != nil || len(storedID) != len(route.PublicID) {
		_ = database.Rollback()
		if err != nil {
			return AnalyticsPublicRoute{}, err
		}
		return AnalyticsPublicRoute{}, fmt.Errorf("analytics public route result is invalid")
	}
	copy(route.PublicID[:], storedID)
	if err = database.Commit(ctx); err != nil {
		return AnalyticsPublicRoute{}, err
	}
	return route, nil
}

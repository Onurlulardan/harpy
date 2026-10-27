package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type AnalyticsLegacyPublicKey struct {
	TenantIdentity string
	Method         string
	Path           string
	Slug           string
}

type AnalyticsLegacyPublicRoute struct {
	Key                         AnalyticsLegacyPublicKey
	TenantID                    uuid.UUID
	LegacyEndpointID            uuid.UUID
	PublicID                    [16]byte
	ReleaseID                   uuid.UUID
	State                       string
	HitCount                    uint64
	LastHitAt, ConsumersAckedAt *time.Time
	RemoveAfter                 time.Time
}

func (pool *Pool) ResolveAnalyticsLegacyPublic(
	ctx context.Context, key AnalyticsLegacyPublicKey,
) (AnalyticsLegacyPublicRoute, error) {
	if !validAnalyticsLegacyKey(key) {
		return AnalyticsLegacyPublicRoute{}, fmt.Errorf("Analytics legacy key is invalid")
	}
	database, err := pool.BeginRuntimeTransaction(ctx, TransactionSnapshotQuery)
	if err != nil {
		return AnalyticsLegacyPublicRoute{}, err
	}
	route, err := scanAnalyticsLegacyPublic(database.transaction.QueryRow(ctx,
		resolveAnalyticsLegacyPublicSQL, key.TenantIdentity, key.Method, key.Path, key.Slug))
	if err != nil {
		_ = database.Rollback()
		return AnalyticsLegacyPublicRoute{}, err
	}
	if err = database.Commit(ctx); err != nil {
		return AnalyticsLegacyPublicRoute{}, err
	}
	return route, nil
}

func (pool *Pool) RecordAnalyticsLegacyPublicHit(
	ctx context.Context, route AnalyticsLegacyPublicRoute, at time.Time,
) (AnalyticsLegacyPublicRoute, error) {
	if !validAnalyticsLegacyRoute(route) || at.IsZero() || at.Location() != time.UTC {
		return AnalyticsLegacyPublicRoute{}, fmt.Errorf("Analytics legacy hit is invalid")
	}
	database, err := pool.BeginRuntimeTransaction(ctx, TransactionCommand)
	if err != nil {
		return AnalyticsLegacyPublicRoute{}, err
	}
	key := route.Key
	result, err := scanAnalyticsLegacyPublic(database.transaction.QueryRow(ctx,
		recordAnalyticsLegacyPublicHitSQL, key.TenantIdentity, key.Method, key.Path,
		key.Slug, route.TenantID, route.LegacyEndpointID, route.PublicID[:],
		route.ReleaseID, at))
	if err != nil {
		_ = database.Rollback()
		return AnalyticsLegacyPublicRoute{}, err
	}
	if err = database.Commit(ctx); err != nil {
		return AnalyticsLegacyPublicRoute{}, err
	}
	return result, nil
}

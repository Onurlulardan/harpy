package api

import (
	"context"
	"fmt"
	"time"

	analyticsmodule "github.com/Onurlulardan/harpy/core/internal/modules/analytics"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
)

type analyticsLegacyPublicDatabase struct{ pool *platformpostgres.Pool }

func (database analyticsLegacyPublicDatabase) ResolveAnalyticsLegacyPublic(
	ctx context.Context, key analyticsmodule.LegacyPublicKey,
) (analyticsmodule.LegacyPublicMapping, error) {
	if database.pool == nil || ctx == nil || ctx.Err() != nil {
		return analyticsmodule.LegacyPublicMapping{},
			fmt.Errorf("Analytics legacy public database is unavailable")
	}
	result, err := database.pool.ResolveAnalyticsLegacyPublic(ctx,
		platformpostgres.AnalyticsLegacyPublicKey{
			TenantIdentity: key.TenantIdentity, Method: key.Method,
			Path: key.Path, Slug: key.Slug,
		})
	if err != nil {
		return analyticsmodule.LegacyPublicMapping{}, err
	}
	return analyticsLegacyPublicMapping(result), nil
}

func (database analyticsLegacyPublicDatabase) RecordAnalyticsLegacyPublicHit(
	ctx context.Context, mapping analyticsmodule.LegacyPublicMapping, at time.Time,
) (analyticsmodule.LegacyPublicMapping, error) {
	if database.pool == nil || ctx == nil || ctx.Err() != nil {
		return analyticsmodule.LegacyPublicMapping{},
			fmt.Errorf("Analytics legacy public database is unavailable")
	}
	result, err := database.pool.RecordAnalyticsLegacyPublicHit(
		ctx, platformAnalyticsLegacyPublicRoute(mapping), at,
	)
	if err != nil {
		return analyticsmodule.LegacyPublicMapping{}, err
	}
	return analyticsLegacyPublicMapping(result), nil
}

func platformAnalyticsLegacyPublicRoute(
	value analyticsmodule.LegacyPublicMapping,
) platformpostgres.AnalyticsLegacyPublicRoute {
	return platformpostgres.AnalyticsLegacyPublicRoute{
		Key: platformpostgres.AnalyticsLegacyPublicKey{
			TenantIdentity: value.Key.TenantIdentity, Method: value.Key.Method,
			Path: value.Key.Path, Slug: value.Key.Slug,
		},
		TenantID: value.TenantID, LegacyEndpointID: value.LegacyEndpointID,
		PublicID: value.PublicID, ReleaseID: value.ReleaseID, State: value.State,
		HitCount: value.HitCount, LastHitAt: value.LastHitAt,
		ConsumersAckedAt: value.ConsumersAckedAt, RemoveAfter: value.RemoveAfter,
	}
}

func analyticsLegacyPublicMapping(
	value platformpostgres.AnalyticsLegacyPublicRoute,
) analyticsmodule.LegacyPublicMapping {
	return analyticsmodule.LegacyPublicMapping{
		Key: analyticsmodule.LegacyPublicKey{
			TenantIdentity: value.Key.TenantIdentity, Method: value.Key.Method,
			Path: value.Key.Path, Slug: value.Key.Slug,
		},
		TenantID: value.TenantID, LegacyEndpointID: value.LegacyEndpointID,
		PublicID: value.PublicID, ReleaseID: value.ReleaseID, State: value.State,
		HitCount: value.HitCount, LastHitAt: value.LastHitAt,
		ConsumersAckedAt: value.ConsumersAckedAt, RemoveAfter: value.RemoveAfter,
	}
}

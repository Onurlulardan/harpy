package postgres

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

func scanAnalyticsLegacyPublic(row Row) (AnalyticsLegacyPublicRoute, error) {
	var route AnalyticsLegacyPublicRoute
	var publicID []byte
	err := row.Scan(&route.Key.TenantIdentity, &route.Key.Method, &route.Key.Path,
		&route.Key.Slug, &route.TenantID, &route.LegacyEndpointID, &publicID,
		&route.ReleaseID, &route.State, &route.HitCount, &route.LastHitAt,
		&route.ConsumersAckedAt, &route.RemoveAfter)
	if err != nil {
		return AnalyticsLegacyPublicRoute{}, err
	}
	if len(publicID) != len(route.PublicID) {
		return AnalyticsLegacyPublicRoute{}, fmt.Errorf("Analytics legacy route is invalid")
	}
	copy(route.PublicID[:], publicID)
	if !validAnalyticsLegacyRoute(route) {
		return AnalyticsLegacyPublicRoute{}, fmt.Errorf("Analytics legacy route is invalid")
	}
	return route, nil
}

func validAnalyticsLegacyKey(key AnalyticsLegacyPublicKey) bool {
	return key.TenantIdentity != "" && strings.TrimSpace(key.TenantIdentity) == key.TenantIdentity &&
		(key.Method == "GET" || key.Method == "POST") && strings.HasPrefix(key.Path, "/") &&
		len(key.Path) <= 256 && key.Slug != "" && strings.TrimSpace(key.Slug) == key.Slug
}

func validAnalyticsLegacyRoute(route AnalyticsLegacyPublicRoute) bool {
	return validAnalyticsLegacyKey(route.Key) && route.TenantID != uuid.Nil &&
		route.LegacyEndpointID != uuid.Nil && route.ReleaseID != uuid.Nil &&
		route.PublicID != [16]byte{} && route.State != "" && !route.RemoveAfter.IsZero()
}

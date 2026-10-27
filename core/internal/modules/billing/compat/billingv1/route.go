package billingv1

import "net/http"

type MappingPolicy string

const (
	MappingPreserved  MappingPolicy = "preserved"
	MappingVersion    MappingPolicy = "immutable_version"
	MappingRetire     MappingPolicy = "retire"
	MappingTransition MappingPolicy = "explicit_transition"
)

type Route struct {
	ID, Method, LegacyPath, RuntimePattern string
	Target, Permission                     string
	Policy                                 MappingPolicy
}

type RouteBinding struct {
	Method, Pattern string
	Handler         http.Handler
}

func Routes() []Route {
	routes := append([]Route{}, publicTenantRoutes...)
	routes = append(routes, adminPlanRoutes...)
	routes = append(routes, adminReportRoutes...)
	return append(routes, adminPromotionRoutes...)
}

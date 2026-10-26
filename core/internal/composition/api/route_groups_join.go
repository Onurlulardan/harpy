package api

import "github.com/Onurlulardan/harpy/core/internal/platform/httpedge"

func joinAPIRouteGroups(parts ...[][]httpedge.Route) [][]httpedge.Route {
	count := 0
	for _, part := range parts {
		count += len(part)
	}
	groups := make([][]httpedge.Route, 0, count)
	for _, part := range parts {
		groups = append(groups, part...)
	}
	return groups
}

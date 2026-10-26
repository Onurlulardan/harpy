package httpedge

type routeMatch struct {
	route  compiledRoute
	values []string
}

func matchRoute(route compiledRoute, requestPath string) (routeMatch, bool) {
	parts := splitRoutePath(requestPath)
	if len(parts) != len(route.segments) {
		return routeMatch{}, false
	}
	values := make([]string, len(parts))
	for index, segment := range route.segments {
		if segment.literal != "" && segment.literal != parts[index] {
			return routeMatch{}, false
		}
		if segment.parameter != "" {
			if parts[index] == "" {
				return routeMatch{}, false
			}
			values[index] = parts[index]
		}
	}
	return routeMatch{route: route, values: values}, true
}

func routesOverlap(left, right compiledRoute) bool {
	if len(left.segments) != len(right.segments) {
		return false
	}
	for index := range left.segments {
		leftLiteral := left.segments[index].literal
		rightLiteral := right.segments[index].literal
		if leftLiteral != "" && rightLiteral != "" && leftLiteral != rightLiteral {
			return false
		}
	}
	return true
}

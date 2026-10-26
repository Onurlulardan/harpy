package workhttp

import "net/http"

type Route struct {
	Method, Pattern string
	Handler         http.Handler
}

type routeSpec struct {
	method, pattern, operation, authority, resource string
	list, compatibility, created                    bool
}

func (handler *Handler) Routes() []Route {
	if !handler.Ready() {
		return nil
	}
	specifications := append(canonicalRoutes(), compatibilityRoutes()...)
	routes := make([]Route, 0, len(specifications))
	for _, specification := range specifications {
		specification := specification
		routes = append(routes, Route{Method: specification.method,
			Pattern: specification.pattern,
			Handler: http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				handler.serve(writer, request, specification)
			})})
	}
	return routes
}

func read(method, pattern, operation, authority, resource string, list bool) routeSpec {
	return routeSpec{method: method, pattern: pattern, operation: operation,
		authority: authority, resource: resource, list: list}
}

func write(method, pattern, operation, authority, resource string, created bool) routeSpec {
	return routeSpec{method: method, pattern: pattern, operation: operation,
		authority: authority, resource: resource, created: created}
}

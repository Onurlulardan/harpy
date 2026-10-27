package cataloghttp

import "net/http"

type Route struct {
	Method, Pattern string
	Handler         http.Handler
}

func NewRoute(method, pattern string, serve http.HandlerFunc) Route {
	return Route{Method: method, Pattern: pattern, Handler: serve}
}

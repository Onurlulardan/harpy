package queryhttp

func routeExists(method, pattern string) bool {
	for _, route := range (&Handler{}).Routes() {
		if route.Method == method && route.Pattern == pattern && route.Handler != nil {
			return true
		}
	}
	return false
}

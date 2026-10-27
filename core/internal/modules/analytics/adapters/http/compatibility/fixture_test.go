package compatibilityhttp

import "testing"

func requireRoute(t *testing.T, method, pattern string) {
	t.Helper()
	for _, route := range (&Handler{}).Routes() {
		if route.Method == method && route.Pattern == pattern && route.Handler != nil {
			return
		}
	}
	t.Fatalf("missing analytics compatibility route %s %s", method, pattern)
}

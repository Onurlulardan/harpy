package requestmeta

import (
	"context"
	"testing"
)

func TestRoutePatternContextRoundTrip(t *testing.T) {
	if _, found := RoutePatternFromContext(context.Background()); found {
		t.Fatal("unexpected route pattern")
	}
	ctx := WithRoutePattern(context.Background(), "/api/users/{user_id}")
	pattern, found := RoutePatternFromContext(ctx)
	if !found || pattern != "/api/users/{user_id}" {
		t.Fatalf("route pattern = %q, %v", pattern, found)
	}
}

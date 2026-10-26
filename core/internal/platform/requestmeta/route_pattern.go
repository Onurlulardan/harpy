package requestmeta

import "context"

type routePatternContextKey struct{}

func WithRoutePattern(ctx context.Context, pattern string) context.Context {
	return context.WithValue(ctx, routePatternContextKey{}, pattern)
}

func RoutePatternFromContext(ctx context.Context) (string, bool) {
	pattern, found := ctx.Value(routePatternContextKey{}).(string)
	return pattern, found && pattern != ""
}

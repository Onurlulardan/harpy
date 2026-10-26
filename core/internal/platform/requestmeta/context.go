package requestmeta

import "context"

type correlationContextKey struct{}

func WithCorrelationID(ctx context.Context, id CorrelationID) context.Context {
	return context.WithValue(ctx, correlationContextKey{}, id)
}

func CorrelationIDFromContext(ctx context.Context) (CorrelationID, bool) {
	id, found := ctx.Value(correlationContextKey{}).(CorrelationID)
	return id, found && id != ""
}

package authority

import "context"

type contextKey struct{}

func WithEvidence(ctx context.Context, evidence Evidence) (context.Context, error) {
	if ctx == nil || ctx.Err() != nil || !evidence.valid() {
		return nil, ErrInvalid
	}
	return context.WithValue(ctx, contextKey{}, evidence), nil
}

func FromContext(ctx context.Context) (Evidence, bool) {
	if ctx == nil || ctx.Err() != nil {
		return Evidence{}, false
	}
	evidence, ok := ctx.Value(contextKey{}).(Evidence)
	return evidence, ok && evidence.valid()
}

package csrf

import "context"

type KeyResolver interface {
	CSRFKey(context.Context, string) ([]byte, error)
}

type KeyResolverFunc func(context.Context, string) ([]byte, error)

func (function KeyResolverFunc) CSRFKey(
	ctx context.Context,
	keyID string,
) ([]byte, error) {
	return function(ctx, keyID)
}

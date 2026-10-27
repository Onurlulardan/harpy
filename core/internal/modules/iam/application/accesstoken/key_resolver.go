package accesstoken

import (
	"context"
	"crypto/ed25519"
)

type KeyResolver interface {
	AccessTokenKey(context.Context, string) (ed25519.PublicKey, error)
}

type KeyResolverFunc func(context.Context, string) (ed25519.PublicKey, error)

func (function KeyResolverFunc) AccessTokenKey(
	ctx context.Context,
	keyID string,
) (ed25519.PublicKey, error) {
	return function(ctx, keyID)
}

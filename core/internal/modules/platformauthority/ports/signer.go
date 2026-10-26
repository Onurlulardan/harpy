package ports

import (
	"context"
	"crypto/ed25519"
)

type ProposalSigner interface {
	KID() string
	PublicKey(context.Context) (ed25519.PublicKey, error)
	Sign(context.Context, []byte) ([]byte, error)
}

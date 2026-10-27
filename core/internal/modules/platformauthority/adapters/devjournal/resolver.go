package devjournal

import (
	"context"
	"crypto/ed25519"
)

type ProposalKeyResolver interface {
	ResolveProposalKey(context.Context, string) (ed25519.PublicKey, error)
}

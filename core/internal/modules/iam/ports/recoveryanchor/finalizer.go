package recoveryanchor

import "context"

type AuthorityFinalizerStore interface {
	FinalizeAuthority(context.Context, string) error
}

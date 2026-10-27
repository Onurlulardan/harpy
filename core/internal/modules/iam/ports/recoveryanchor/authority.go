package recoveryanchor

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/recoveryanchor"
)

type AuthorityPlanner interface {
	Prepare(context.Context, recoveryanchor.ID, recoveryanchor.AuthorityState, time.Time) (
		recoveryanchor.PreparedAuthority, error)
}

type AuthorityStore[S any] interface {
	PersistProposal(context.Context, S, recoveryanchor.Subject,
		recoveryanchor.PreparedAuthority, time.Time) error
	Commit(context.Context, S, recoveryanchor.Subject, recoveryanchor.ID, time.Time) (
		recoveryanchor.CommitResult, error)
}

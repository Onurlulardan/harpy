package platformreport

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type TransitionTargetStore interface {
	ResolveTransitionTarget(context.Context, uuid.UUID) (TransitionTarget, error)
}

type TransitionTargetBinder interface {
	BindTransitionTarget(
		platformtx.GlobalScope, platformtx.OwnerAuthority,
	) (TransitionTargetStore, error)
}

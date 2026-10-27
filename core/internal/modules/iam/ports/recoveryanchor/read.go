package recoveryanchor

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/recoveryanchor"
)

type Reader[S any] interface {
	List(context.Context, S, recoveryanchor.Subject) ([]recoveryanchor.Anchor, error)
}

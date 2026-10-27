package recoveryanchor

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/recoveryanchor"
)

type DeleteStore[S any] interface {
	StageDelete(context.Context, S, recoveryanchor.DeleteCommand) (
		recoveryanchor.StagedMutation, error)
}

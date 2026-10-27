package recoveryanchor

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/recoveryanchor"
)

type RegistrationStore[S any] interface {
	StageRegistration(context.Context, S, recoveryanchor.RegisterCommand) (
		recoveryanchor.StagedMutation, error)
}

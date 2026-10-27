package recoveryanchor

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/recoveryanchor"
)

type OptionsStore[S any] interface {
	CreateOptions(context.Context, S, recoveryanchor.OptionsCommand) (
		recoveryanchor.RegistrationOptions, error)
	LoadRegistration(context.Context, S, recoveryanchor.Subject, string) (
		recoveryanchor.RegistrationLookup, error)
}

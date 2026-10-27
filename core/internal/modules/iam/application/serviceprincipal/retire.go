package serviceprincipal

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (commands *Commands) Retire(
	ctx context.Context, scope transaction.TenantScope, command StateCommand,
) (MutationResult, error) {
	if commands == nil {
		return MutationResult{}, ErrInvalidPrincipal
	}
	return commands.changeState(
		ctx, scope, command, commands.port.Retire,
		"iam.service_principal.retire", "iam.service_principal.retired",
	)
}

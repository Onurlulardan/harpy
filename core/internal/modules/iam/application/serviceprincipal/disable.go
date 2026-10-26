package serviceprincipal

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (commands *Commands) Disable(
	ctx context.Context, scope transaction.TenantScope, command StateCommand,
) (MutationResult, error) {
	if commands == nil {
		return MutationResult{}, ErrInvalidPrincipal
	}
	return commands.changeState(
		ctx, scope, command, commands.port.Disable,
		"iam.service_principal.disable", "iam.service_principal.disabled",
	)
}

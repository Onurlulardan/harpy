package serviceprincipal

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (commands *Commands) Enable(
	ctx context.Context, scope transaction.TenantScope, command StateCommand,
) (MutationResult, error) {
	if commands == nil {
		return MutationResult{}, ErrInvalidPrincipal
	}
	return commands.changeState(
		ctx, scope, command, commands.port.Enable,
		"iam.service_principal.enable", "iam.service_principal.enabled",
	)
}

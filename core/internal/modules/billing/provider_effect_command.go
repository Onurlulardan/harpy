package billing

import (
	"context"

	effectsv1 "github.com/Onurlulardan/harpy/contracts/effects/v1"
	providerapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/provider"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/effectexecution"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type ProviderEffectCommand struct {
	identity effectexecution.Identity
	executor providerapp.EffectExecutor
}

func NewProviderEffectCommand(
	registry providerapp.Registry,
	providerName string,
	operation string,
) (*ProviderEffectCommand, error) {
	executor, err := registry.EffectFor(providerName)
	if err != nil {
		return nil, err
	}
	identity, err := effectexecution.NewIdentity(providerName, operation)
	if err != nil {
		return nil, providerapp.ErrConfiguration
	}
	return &ProviderEffectCommand{identity: identity, executor: executor}, nil
}

func (command *ProviderEffectCommand) Identity() effectexecution.Identity {
	if command == nil {
		return effectexecution.Identity{}
	}
	return command.identity
}

func (command *ProviderEffectCommand) Execute(
	ctx context.Context,
	request effectsv1.Request,
) effectexecution.Result {
	if command == nil || command.executor == nil || !transaction.NetworkAllowed(ctx) {
		return effectexecution.Result{}
	}
	result, err := command.executor.ExecuteEffect(ctx, request)
	if err != nil || !result.ValidFor(request) {
		return effectexecution.Result{}
	}
	return providerPlatformResult(result)
}

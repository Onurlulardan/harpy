package effectexecution

import (
	"context"
	"errors"
	"fmt"

	effectsv1 "github.com/Onurlulardan/harpy/contracts/effects/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

var (
	ErrProviderNotRegistered = errors.New("effect provider operation is not registered")
	ErrNetworkInsideScope    = errors.New("external effect cannot run inside TenantScope")
	ErrInvocationContext     = errors.New("effect provider invocation context is invalid")
	ErrProviderResult        = errors.New("effect provider returned an invalid result")
)

type Provider interface {
	Identity() Identity
	Execute(context.Context, effectsv1.Request) Result
}

type Registry struct {
	providers map[string]Provider
}

func NewRegistry(providers []Provider) (Registry, error) {
	registered := make(map[string]Provider, len(providers))
	for index, provider := range providers {
		if provider == nil || !provider.Identity().valid() {
			return Registry{}, fmt.Errorf("effect provider %d is invalid", index)
		}
		key := provider.Identity().key()
		if _, exists := registered[key]; exists {
			return Registry{}, fmt.Errorf("duplicate effect provider operation")
		}
		registered[key] = provider
	}
	return Registry{providers: registered}, nil
}

func (registry Registry) Resolve(identity Identity) (Resolved, error) {
	if !identity.valid() || registry.providers == nil {
		return Resolved{}, ErrProviderNotRegistered
	}
	provider, exists := registry.providers[identity.key()]
	if !exists {
		return Resolved{}, ErrProviderNotRegistered
	}
	return Resolved{identity: identity, provider: provider}, nil
}

type Resolved struct {
	identity Identity
	provider Provider
}

func (resolved Resolved) Invoke(
	ctx context.Context,
	request effectsv1.Request,
) (Result, error) {
	if ctx == nil || ctx.Err() != nil {
		return Result{}, ErrInvocationContext
	}
	if !transaction.NetworkAllowed(ctx) {
		return Result{}, ErrNetworkInsideScope
	}
	if resolved.provider == nil || resolved.provider.Identity() != resolved.identity ||
		request.Provider != resolved.identity.provider ||
		request.Operation != resolved.identity.operation || request.Validate() != nil {
		return Result{}, ErrProviderNotRegistered
	}
	original, err := cloneRequest(request)
	if err != nil {
		return Result{}, ErrProviderNotRegistered
	}
	providerRequest, err := cloneRequest(request)
	if err != nil {
		return Result{}, ErrProviderNotRegistered
	}
	result := resolved.provider.Execute(ctx, providerRequest)
	if err := result.validateFor(original); err != nil {
		return Result{}, fmt.Errorf("%w: %v", ErrProviderResult, err)
	}
	return result, nil
}

package execution

import (
	"context"

	restqueryapi "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func MintProgram(
	ctx context.Context,
	scope transaction.TenantScope,
	reference restqueryapi.CatalogReference,
	execute func(context.Context, SnapshotDriver) error,
) (Program, error) {
	identity, err := captureScopeIdentity(ctx, scope)
	if err != nil || reference.Validate() != nil {
		return Program{}, errExecutionContract
	}
	owner, err := reference.Owner()
	if err != nil || transaction.RequireOwner(scope, owner) != nil {
		return Program{}, errExecutionContract
	}
	ownedContext, cancel := context.WithCancel(ctx)
	program, err := newProgram(
		ownedContext, cancel, identity, owner, execute, verifyCapturedScope,
	)
	if err != nil {
		cancel()
		return Program{}, errExecutionContract
	}
	return program, nil
}

func newProgram(
	ctx context.Context,
	cancel context.CancelFunc,
	identity scopeIdentity,
	owner string,
	execute func(context.Context, SnapshotDriver) error,
	verify scopeVerifier,
) (Program, error) {
	if ctx == nil || cancel == nil || ctx.Err() != nil || !identity.valid() ||
		!ownerNamePattern.MatchString(owner) || execute == nil || verify == nil {
		return Program{}, errExecutionContract
	}
	state := &programState{
		context: ctx, cancel: cancel, identity: identity,
		owner: owner, execute: execute, verify: verify,
	}
	state.phase.Store(uint32(programReady))
	return Program{state: state}, nil
}

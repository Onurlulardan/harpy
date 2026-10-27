package execution

import (
	"context"
	"fmt"
	"log/slog"
	"sync/atomic"

	queryapi "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type globalProgramState struct {
	phase    atomic.Uint32
	context  context.Context
	cancel   context.CancelFunc
	identity globalScopeIdentity
	owner    string
	execute  func(context.Context, GlobalSnapshotDriver) error
	outcome  error
}

type GlobalProgram struct{ state *globalProgramState }

func MintGlobalProgram(
	ctx context.Context,
	scope transaction.GlobalScope,
	reference queryapi.CatalogReference,
	execute func(context.Context, GlobalSnapshotDriver) error,
) (GlobalProgram, error) {
	identity, err := captureGlobalScopeIdentity(ctx, scope)
	if err != nil || reference.Validate() != nil || execute == nil {
		return GlobalProgram{}, errExecutionContract
	}
	owner, err := reference.Owner()
	if err != nil || transaction.RequireGlobalOwner(scope, owner) != nil {
		return GlobalProgram{}, errExecutionContract
	}
	ownedContext, cancel := context.WithCancel(ctx)
	state := &globalProgramState{
		context: ownedContext, cancel: cancel, identity: identity,
		owner: owner, execute: execute,
	}
	state.phase.Store(uint32(programReady))
	program := GlobalProgram{state: state}
	if program.Validate() != nil {
		cancel()
		return GlobalProgram{}, errExecutionContract
	}
	return program, nil
}

func (program GlobalProgram) Validate() error {
	if program.state == nil || program.state.context == nil || program.state.cancel == nil ||
		!program.state.identity.valid() || !ownerNamePattern.MatchString(program.state.owner) ||
		program.state.execute == nil {
		return errExecutionContract
	}
	phase := programPhase(program.state.phase.Load())
	if phase < programReady || phase > programCompleted {
		return errExecutionContract
	}
	return nil
}

func (GlobalProgram) String() string                         { return executionRedaction }
func (GlobalProgram) GoString() string                       { return executionRedaction }
func (GlobalProgram) LogValue() slog.Value                   { return executionLogValue() }
func (program GlobalProgram) Format(state fmt.State, _ rune) { writeExecutionRedaction(state) }
func (GlobalProgram) MarshalJSON() ([]byte, error)           { return rejectExecutionJSON() }

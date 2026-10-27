package execution

import (
	"context"
	"fmt"
	"log/slog"
	"sync/atomic"
)

type programPhase uint32

const (
	programReady programPhase = iota + 1
	programDispatched
	programRunning
	programCompleted
)

type scopeVerifier func(context.Context, scopeIdentity) error

type programState struct {
	phase    atomic.Uint32
	context  context.Context
	cancel   context.CancelFunc
	identity scopeIdentity
	owner    string
	execute  func(context.Context, SnapshotDriver) error
	verify   scopeVerifier
	outcome  error
}

type Program struct {
	state *programState
}

func (program Program) Validate() error {
	if program.state == nil || program.state.context == nil ||
		program.state.cancel == nil || !program.state.identity.valid() ||
		!ownerNamePattern.MatchString(program.state.owner) ||
		program.state.execute == nil || program.state.verify == nil {
		return errExecutionContract
	}
	phase := programPhase(program.state.phase.Load())
	if phase < programReady || phase > programCompleted {
		return errExecutionContract
	}
	return nil
}

func (Program) String() string   { return executionRedaction }
func (Program) GoString() string { return executionRedaction }
func (Program) LogValue() slog.Value {
	return executionLogValue()
}
func (program Program) Format(state fmt.State, _ rune) {
	writeExecutionRedaction(state)
}
func (Program) MarshalJSON() ([]byte, error) { return rejectExecutionJSON() }

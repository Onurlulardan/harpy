package execution

func (program GlobalProgram) beginDispatch() error {
	if program.Validate() != nil || !program.state.phase.CompareAndSwap(
		uint32(programReady), uint32(programDispatched),
	) || program.state.context.Err() != nil {
		return errExecutionContract
	}
	return nil
}

func (program GlobalProgram) cancelDispatch() {
	if program.state != nil && program.state.cancel != nil {
		program.state.cancel()
	}
}

func (program GlobalProgram) failBeforeRun() {
	if program.state == nil {
		return
	}
	program.state.outcome = errExecutionContract
	program.state.phase.CompareAndSwap(uint32(programDispatched), uint32(programCompleted))
}

func (program GlobalProgram) finishDispatch() error {
	if program.Validate() != nil ||
		programPhase(program.state.phase.Load()) != programCompleted ||
		program.state.outcome != nil {
		return programOutcome(program.state)
	}
	return nil
}

func programOutcome(state *globalProgramState) error {
	if state == nil || state.outcome == nil {
		return errExecutionContract
	}
	return state.outcome
}

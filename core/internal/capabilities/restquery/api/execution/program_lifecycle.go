package execution

func (program Program) beginDispatch() error {
	if program.Validate() != nil || !program.state.phase.CompareAndSwap(
		uint32(programReady), uint32(programDispatched),
	) {
		return errExecutionContract
	}
	return nil
}

func (program Program) failBeforeRun() {
	if program.state == nil || !program.state.phase.CompareAndSwap(
		uint32(programDispatched), uint32(programRunning),
	) {
		return
	}
	program.state.outcome = errExecutionContract
	program.state.phase.Store(uint32(programCompleted))
}

func (program Program) finishDispatch() error {
	if program.state == nil {
		return errExecutionContract
	}
	phase := programPhase(program.state.phase.Load())
	if phase != programCompleted {
		if phase == programDispatched {
			program.failBeforeRun()
		}
		program.state.cancel()
		return errExecutionContract
	}
	outcome := program.state.outcome
	program.state.cancel()
	return outcome
}

func (program Program) cancelDispatch() {
	if program.state != nil && program.state.cancel != nil {
		program.state.cancel()
	}
}

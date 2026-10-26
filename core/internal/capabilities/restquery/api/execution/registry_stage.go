package execution

type stageOutcome struct {
	err        error
	panicValue any
	index      int
}

// RunStage preflights every Program before dispatching the same-snapshot set.
func (registry Registry) RunStage(stage ProgramStage) error {
	if registry.Validate() != nil || stage.validateReady() != nil ||
		registry.preflightStage(stage) != nil {
		failStageBeforeRun(stage.programs)
		return errExecutionContract
	}
	for index := range stage.programs {
		if stage.programs[index].beginDispatch() != nil {
			failStageBeforeRun(stage.programs)
			return errExecutionContract
		}
	}
	outcomes := make(chan stageOutcome, len(stage.programs))
	for index, program := range stage.programs {
		port := registry.ports[program.state.owner]
		go dispatchStageProgram(index, program, port, outcomes)
	}
	results := make([]stageOutcome, len(stage.programs))
	for range stage.programs {
		outcome := <-outcomes
		results[outcome.index] = outcome
		if outcome.err != nil || outcome.panicValue != nil {
			cancelStage(stage.programs)
		}
	}
	return resolveStageOutcome(results)
}

func (registry Registry) preflightStage(stage ProgramStage) error {
	for _, program := range stage.programs {
		if program.state.verify(
			program.state.context, program.state.identity,
		) != nil {
			return errExecutionContract
		}
		port, found := registry.ports[program.state.owner]
		if !found || nilInterface(port) {
			return errExecutionContract
		}
	}
	return nil
}

func dispatchStageProgram(
	index int,
	program Program,
	port OwnerPort,
	outcomes chan<- stageOutcome,
) {
	outcome := stageOutcome{index: index}
	var portErr error
	defer func() {
		outcome.panicValue = recover()
		outcome.err = mergePortOutcome(portErr, program.finishDispatch())
		outcomes <- outcome
	}()
	portErr = port.Run(program.state.identity.scope, program)
}

func failStageBeforeRun(programs []Program) {
	for _, program := range programs {
		if program.state == nil {
			continue
		}
		phase := programPhase(program.state.phase.Load())
		if phase == programReady && program.state.phase.CompareAndSwap(
			uint32(programReady), uint32(programDispatched),
		) {
			program.failBeforeRun()
		} else if phase == programDispatched {
			program.failBeforeRun()
		}
		program.cancelDispatch()
	}
}

func cancelStage(programs []Program) {
	for _, program := range programs {
		program.cancelDispatch()
	}
}

package execution

func (registry Registry) Run(program Program) error {
	if registry.Validate() != nil || program.beginDispatch() != nil {
		return errExecutionContract
	}
	defer program.cancelDispatch()
	if program.state.verify(
		program.state.context, program.state.identity,
	) != nil {
		program.failBeforeRun()
		return program.finishDispatch()
	}
	port, found := registry.ports[program.state.owner]
	if !found || nilInterface(port) {
		program.failBeforeRun()
		return program.finishDispatch()
	}
	portErr := port.Run(program.state.identity.scope, program)
	return mergePortOutcome(portErr, program.finishDispatch())
}

package execution

func (program GlobalProgram) Run(driver GlobalSnapshotDriver) (result error) {
	if program.state == nil || !program.state.phase.CompareAndSwap(
		uint32(programDispatched), uint32(programRunning),
	) {
		return errExecutionContract
	}
	returned := false
	var lease *driverLease
	defer func() {
		if lease != nil {
			lease.close()
		}
		if returned {
			program.state.outcome = result
		} else {
			program.state.outcome = errExecutionContract
		}
		program.state.phase.Store(uint32(programCompleted))
	}()
	if verifyCapturedGlobalScope(program.state.context, program.state.identity) != nil {
		result, returned = errExecutionContract, true
		return result
	}
	var err error
	lease, err = newGlobalDriverLease(
		program.state.context, driver, program.state.owner, program.state.identity.scope,
	)
	if err != nil {
		result, returned = errExecutionContract, true
		return result
	}
	result = program.state.execute(program.state.context, leasedGlobalDriver{lease: lease})
	returned = true
	return result
}

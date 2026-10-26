package execution

import "github.com/Onurlulardan/harpy/core/internal/platform/transaction"

type GlobalOwnerPort interface {
	Run(transaction.GlobalScope, GlobalProgram) error
}

type GlobalRegistration struct {
	owner string
	port  GlobalOwnerPort
}

func NewGlobalRegistration(owner string, port GlobalOwnerPort) (GlobalRegistration, error) {
	if !ownerNamePattern.MatchString(owner) || nilInterface(port) {
		return GlobalRegistration{}, errExecutionContract
	}
	return GlobalRegistration{owner: owner, port: port}, nil
}

type GlobalRegistry struct {
	ports map[string]GlobalOwnerPort
}

func NewGlobalRegistry(registrations []GlobalRegistration) (GlobalRegistry, error) {
	if len(registrations) == 0 {
		return GlobalRegistry{}, errExecutionContract
	}
	ports := make(map[string]GlobalOwnerPort, len(registrations))
	for _, registration := range registrations {
		if !ownerNamePattern.MatchString(registration.owner) || nilInterface(registration.port) {
			return GlobalRegistry{}, errExecutionContract
		}
		if _, duplicate := ports[registration.owner]; duplicate {
			return GlobalRegistry{}, errExecutionContract
		}
		ports[registration.owner] = registration.port
	}
	return GlobalRegistry{ports: ports}, nil
}

func (registry GlobalRegistry) Validate() error {
	if len(registry.ports) == 0 {
		return errExecutionContract
	}
	for owner, port := range registry.ports {
		if !ownerNamePattern.MatchString(owner) || nilInterface(port) {
			return errExecutionContract
		}
	}
	return nil
}

func (registry GlobalRegistry) Run(program GlobalProgram) error {
	if registry.Validate() != nil || program.beginDispatch() != nil {
		return errExecutionContract
	}
	defer program.cancelDispatch()
	if verifyCapturedGlobalScope(program.state.context, program.state.identity) != nil {
		program.failBeforeRun()
		return program.finishDispatch()
	}
	port, found := registry.ports[program.state.owner]
	if !found || nilInterface(port) {
		program.failBeforeRun()
		return program.finishDispatch()
	}
	_ = port.Run(program.state.identity.scope, program)
	return program.finishDispatch()
}

package providerapp

type CommandState string

const (
	CommandQueued            CommandState = "queued"
	CommandRunning           CommandState = "running"
	CommandSucceeded         CommandState = "succeeded"
	CommandFailed            CommandState = "failed"
	CommandReconcileRequired CommandState = "reconcile_required"
)

func (state CommandState) Valid() bool {
	switch state {
	case CommandQueued, CommandRunning, CommandSucceeded,
		CommandFailed, CommandReconcileRequired:
		return true
	default:
		return false
	}
}

func (state CommandState) Terminal() bool {
	return state == CommandSucceeded || state == CommandFailed
}

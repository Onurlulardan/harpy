package dashboardexport

import "fmt"

func finishKnownRender(
	spool EncryptedSpool, state State,
) (RenderExecutionResult, error) {
	execution, err := renderExecutionState(state)
	if err != nil {
		return RenderExecutionResult{State: RenderExecutionOutcomeUnknown}, err
	}
	return RenderExecutionResult{State: execution,
		CleanupError: spool.Discard()}, nil
}

func renderExecutionState(state State) (RenderExecutionState, error) {
	switch state {
	case StateReady:
		return RenderExecutionReady, nil
	case StateCancelled, StateExpired:
		return RenderExecutionCancelled, nil
	case StateQueued:
		return RenderExecutionRequeued, nil
	case StateFailed:
		return RenderExecutionDead, nil
	default:
		return "", fmt.Errorf(
			"Analytics dashboard export terminal state %q is invalid", state)
	}
}

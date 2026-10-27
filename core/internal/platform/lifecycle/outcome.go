package lifecycle

import runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"

type Reason string

const (
	ReasonSignal        Reason = "signal"
	ReasonContext       Reason = "context"
	ReasonServeFailure  Reason = "serve_failure"
	ReasonDrainFailure  Reason = "drain_failure"
	ReasonDrainDeadline Reason = "drain_deadline"
	ReasonSecondSignal  Reason = "second_signal"
)

type Outcome struct {
	exitCode runtimev1.ExitCode
	reason   Reason
}

func newOutcome(exitCode runtimev1.ExitCode, reason Reason) Outcome {
	return Outcome{exitCode: exitCode, reason: reason}
}

func (outcome Outcome) ExitCode() runtimev1.ExitCode { return outcome.exitCode }
func (outcome Outcome) Reason() Reason               { return outcome.reason }

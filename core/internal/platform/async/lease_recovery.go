package async

import (
	"fmt"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
)

type LeaseRecovery struct {
	State        jobsv1.State
	Code         string
	ReuseReceipt bool
}

// DecideLeaseRecovery freezes the ADR-0011 crash boundary for an expired lease.
func DecideLeaseRecovery(
	cancelRequested bool,
	effectStarted bool,
	hasEffectReceipt bool,
) (LeaseRecovery, error) {
	if hasEffectReceipt && !effectStarted {
		return LeaseRecovery{}, fmt.Errorf("effect receipt exists before effect start")
	}
	if hasEffectReceipt {
		return LeaseRecovery{State: jobsv1.StateQueued, ReuseReceipt: true}, nil
	}
	if effectStarted {
		return LeaseRecovery{
			State: jobsv1.StateDead,
			Code:  jobsv1.UnknownOutcomeCode,
		}, nil
	}
	if cancelRequested {
		return LeaseRecovery{State: jobsv1.StateCancelled}, nil
	}
	return LeaseRecovery{State: jobsv1.StateQueued}, nil
}

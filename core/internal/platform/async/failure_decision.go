package async

import (
	"fmt"
	"time"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
)

type FailureDecision struct {
	State jobsv1.State
	Code  string
	Delay time.Duration
}

func DecideFailure(
	policy jobsv1.KindPolicy,
	attempt int,
	errorClass jobsv1.ErrorClass,
	errorCode string,
	effectStarted bool,
	hasFailedEffectReceipt bool,
	jitter JitterSource,
) (FailureDecision, error) {
	if err := policy.Validate(); err != nil || attempt < 1 || attempt > policy.MaximumAttempts {
		return FailureDecision{}, fmt.Errorf("job failure policy or attempt is invalid")
	}
	if hasFailedEffectReceipt && !effectStarted {
		return FailureDecision{}, fmt.Errorf("effect receipt exists before effect start")
	}
	if policy.EffectClass == jobsv1.EffectDatabaseTransactional &&
		(effectStarted || hasFailedEffectReceipt) {
		return FailureDecision{}, fmt.Errorf("transactional job has an external effect boundary")
	}
	if effectStarted && !hasFailedEffectReceipt {
		return FailureDecision{
			State: jobsv1.StateDead,
			Code:  jobsv1.UnknownOutcomeCode,
		}, nil
	}
	if errorCode == "" || errorCode == jobsv1.UnknownOutcomeCode {
		return FailureDecision{}, fmt.Errorf("job failure code is invalid")
	}
	switch errorClass {
	case jobsv1.ErrorCancelled:
		return FailureDecision{State: jobsv1.StateCancelled, Code: errorCode}, nil
	case jobsv1.ErrorPermanent:
		return FailureDecision{State: jobsv1.StateDead, Code: errorCode}, nil
	case jobsv1.ErrorTransient, jobsv1.ErrorConflict:
		if attempt == policy.MaximumAttempts {
			return FailureDecision{State: jobsv1.StateDead, Code: errorCode}, nil
		}
		delay, err := RetryDelay(attempt, jitter)
		if err != nil {
			return FailureDecision{}, err
		}
		return FailureDecision{State: jobsv1.StateQueued, Code: errorCode, Delay: delay}, nil
	default:
		return FailureDecision{}, fmt.Errorf("job failure class is invalid")
	}
}

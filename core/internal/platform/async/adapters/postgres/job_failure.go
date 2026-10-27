package postgres

import (
	"regexp"
	"time"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
)

var jobErrorCodePattern = regexp.MustCompile(`^[A-Z][A-Z0-9_]{1,63}$`)

type JobFailure struct {
	errorClass jobsv1.ErrorClass
	code       string
	digest     string
	retryDelay time.Duration
}

func NewJobFailure(
	errorClass jobsv1.ErrorClass,
	code string,
	digest string,
	retryDelay time.Duration,
) (JobFailure, error) {
	failure := JobFailure{
		errorClass: errorClass, code: code, digest: digest, retryDelay: retryDelay,
	}
	validClass := errorClass == jobsv1.ErrorTransient ||
		errorClass == jobsv1.ErrorPermanent || errorClass == jobsv1.ErrorConflict ||
		errorClass == jobsv1.ErrorCancelled
	retryClass := errorClass == jobsv1.ErrorTransient || errorClass == jobsv1.ErrorConflict
	if !validClass || !jobErrorCodePattern.MatchString(code) ||
		code == jobsv1.UnknownOutcomeCode || !claimDigestPattern.MatchString(digest) ||
		retryDelay < 0 || retryDelay > 15*time.Minute || (!retryClass && retryDelay != 0) {
		return JobFailure{}, ErrJobFailureInvalid
	}
	return failure, nil
}

type jobFailureDecision struct {
	state   jobsv1.State
	outcome string
}

func decideJobFailure(claim JobClaim, failure JobFailure) (jobFailureDecision, error) {
	switch failure.errorClass {
	case jobsv1.ErrorCancelled:
		return jobFailureDecision{state: jobsv1.StateCancelled, outcome: "cancelled"}, nil
	case jobsv1.ErrorPermanent:
		return jobFailureDecision{state: jobsv1.StateDead, outcome: "dead"}, nil
	case jobsv1.ErrorTransient, jobsv1.ErrorConflict:
		if claim.attempt >= claim.maximumAttempts {
			if failure.retryDelay != 0 {
				return jobFailureDecision{}, ErrJobFailureInvalid
			}
			return jobFailureDecision{state: jobsv1.StateDead, outcome: "dead"}, nil
		}
		return jobFailureDecision{state: jobsv1.StateQueued, outcome: "requeued"}, nil
	default:
		return jobFailureDecision{}, ErrJobFailureInvalid
	}
}

type JobFailureResult struct {
	state      jobsv1.State
	occurredAt time.Time
	runAt      time.Time
}

func (result JobFailureResult) State() jobsv1.State   { return result.state }
func (result JobFailureResult) OccurredAt() time.Time { return result.occurredAt }
func (result JobFailureResult) RetryAt() (time.Time, bool) {
	return result.runAt, result.state == jobsv1.StateQueued
}

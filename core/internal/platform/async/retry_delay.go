package async

import (
	"fmt"
	"time"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
)

const (
	retryBaseDelay = time.Duration(jobsv1.RetryBaseMilliseconds) * time.Millisecond
	retryDelayCap  = time.Duration(jobsv1.RetryCapMilliseconds) * time.Millisecond
	maximumAttempt = jobsv1.MaximumMaximumAttempts
)

// JitterSource returns a uniformly distributed value in [0, upperBound).
type JitterSource interface {
	Uint64N(upperBound uint64) uint64
}

// RetryDelay implements ADR-0011 full jitter for the failed attempt number.
func RetryDelay(attempt int, source JitterSource) (time.Duration, error) {
	if attempt < 1 || attempt > maximumAttempt || source == nil {
		return 0, fmt.Errorf("async retry delay input is invalid")
	}
	limit := retryBaseDelay
	for current := 1; current < attempt && limit < retryDelayCap; current++ {
		limit *= 2
		if limit > retryDelayCap {
			limit = retryDelayCap
		}
	}
	choice := source.Uint64N(uint64(limit) + 1)
	if choice > uint64(limit) {
		return 0, fmt.Errorf("async jitter source exceeded its bound")
	}
	return time.Duration(choice), nil
}

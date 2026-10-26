package jobsv1

import "time"

func maximumRetryDelay(attempt int) time.Duration {
	maximum := time.Duration(RetryBaseMilliseconds) * time.Millisecond
	cap := time.Duration(RetryCapMilliseconds) * time.Millisecond
	for current := 1; current < attempt && maximum < cap; current++ {
		maximum *= 2
		if maximum > cap {
			maximum = cap
		}
	}
	return maximum
}

func cloneErrorClass(value *ErrorClass) *ErrorClass {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}

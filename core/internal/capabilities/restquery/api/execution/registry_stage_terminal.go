package execution

import (
	"context"
	"errors"
)

func resolveStageOutcome(results []stageOutcome) error {
	for _, result := range results {
		if result.panicValue != nil {
			panic(result.panicValue)
		}
	}
	for _, result := range results {
		if result.err != nil && !isStageCancellation(result.err) {
			return result.err
		}
	}
	for _, result := range results {
		if result.err != nil {
			return result.err
		}
	}
	return nil
}

func isStageCancellation(err error) bool {
	return errors.Is(err, context.Canceled) ||
		errors.Is(err, context.DeadlineExceeded)
}

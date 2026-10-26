package api

import (
	"context"
	"errors"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
)

func newParallelRequestObserver(
	observers ...httpedge.RequestObserver,
) (httpedge.RequestObserver, error) {
	if len(observers) == 0 {
		return nil, fmt.Errorf("API request observers are required")
	}
	for _, observer := range observers {
		if observer == nil {
			return nil, fmt.Errorf("API request observer is unavailable")
		}
	}
	return func(ctx context.Context, observation httpedge.RequestObservation) error {
		results := make(chan error, len(observers))
		for _, observer := range observers {
			current := observer
			go func() { results <- current(ctx, observation) }()
		}
		failures := make([]error, 0, len(observers))
		for range observers {
			select {
			case err := <-results:
				if err != nil {
					failures = append(failures, err)
				}
			case <-ctx.Done():
				failures = append(failures, ctx.Err())
				return errors.Join(failures...)
			}
		}
		return errors.Join(failures...)
	}, nil
}

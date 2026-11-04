package initialize

import (
	"context"
	"errors"
)

func RestartApplication(ctx context.Context, start func(context.Context) error) error {
	if start == nil {
		return &StageError{
			Stage: StageApplication, Phase: "restart", Code: 2,
			Err: errors.New("application restart is not configured"),
		}
	}
	if err := start(ctx); err != nil {
		return &StageError{
			Stage: StageApplication, Phase: "restart", Retryable: true, Code: 6,
			Err: err,
		}
	}
	return nil
}

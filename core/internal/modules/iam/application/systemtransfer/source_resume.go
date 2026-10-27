package systemtransfer

import (
	"context"
	"io"

	tabulartransfer "github.com/Onurlulardan/harpy/core/internal/capabilities/tabulartransfer/api"
)

func (runtime *ImportRuntime) resumeSource(
	ctx context.Context,
	state AttemptState,
	plaintext io.Reader,
) error {
	format, err := tabulartransfer.ParseFormat(state.Source.Format)
	if err != nil || !format.Importable() {
		return ErrInvalidAttempt
	}
	reader, err := tabulartransfer.NewReader(format, plaintext, runtime.limits)
	if err != nil {
		return ErrInvalidAttempt
	}
	cursor, err := NewSourceCursor(reader, format)
	if err != nil || cursor.SkipProcessed(state.Processed) != nil {
		return ErrWorkerState
	}
	for {
		rows, finished, readErr := cursor.NextBatch()
		if readErr != nil {
			return readErr
		}
		if len(rows) > 0 {
			result, processErr := runtime.batches.ProcessBatch(
				ctx, state.Attempt, rows,
			)
			if processErr != nil {
				return processErr
			}
			if result.AlreadyComplete != 0 || result.Committed != len(rows) {
				return ErrWorkerState
			}
		}
		if finished {
			return nil
		}
	}
}

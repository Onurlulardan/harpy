package worker

import (
	"io"
	"log/slog"
)

func reportFailure(output io.Writer, logger *slog.Logger, errorClass string) {
	reportFailureCause(output, logger, errorClass, "")
}

func reportFailureCause(
	output io.Writer, logger *slog.Logger, errorClass, cause string,
) {
	if errorClass == "" || output == nil && logger == nil {
		return
	}
	if logger == nil {
		logger = slog.New(slog.NewJSONHandler(output, &slog.HandlerOptions{Level: slog.LevelError}))
	}
	attrs := []any{
		slog.String("error_class", errorClass),
		slog.String("service", "core-worker"),
	}
	if cause != "" {
		attrs = append(attrs, slog.String("cause", cause))
	}
	logger.Error("core worker failed", attrs...)
}

func reportRuntimeError(logger *slog.Logger) func(error) {
	return func(err error) {
		if logger == nil {
			return
		}
		attrs := []any{
			slog.String("error_class", "runtime"),
			slog.String("service", "core-worker"),
		}
		if err != nil && err.Error() != "" {
			attrs = append(attrs, slog.String("cause", err.Error()))
		}
		logger.Error("core worker runtime error", attrs...)
	}
}

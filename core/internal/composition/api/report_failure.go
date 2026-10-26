package api

import (
	"io"
	"log/slog"
)

func reportFailure(output io.Writer, logger *slog.Logger, errorClass string) {
	if errorClass == "" || output == nil && logger == nil {
		return
	}
	if logger == nil {
		logger = slog.New(slog.NewJSONHandler(output, &slog.HandlerOptions{Level: slog.LevelError}))
	}
	logger.Error(
		"core API failed",
		slog.String("error_class", errorClass),
		slog.String("service", "core-api"),
	)
}

func reportRuntimeError(logger *slog.Logger) func(error) {
	return func(err error) {
		if logger == nil {
			return
		}
		attrs := []any{
			slog.String("error_class", "runtime"),
			slog.String("service", "core-api"),
		}
		if err != nil && err.Error() != "" {
			attrs = append(attrs, slog.String("cause", err.Error()))
		}
		logger.Error("core API runtime error", attrs...)
	}
}

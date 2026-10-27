package migrator

import (
	"io"
	"log/slog"
)

func reportFailure(
	output io.Writer,
	logger *slog.Logger,
	errorClass string,
	command string,
) {
	writeFailureLog(output, logger, errorClass, command, "", "")
}

func reportFailureCause(
	output io.Writer,
	logger *slog.Logger,
	errorClass string,
	command string,
	cause error,
) {
	writeFailureLog(
		output, logger, errorClass, command, "", migrationFailureDetail(cause),
	)
}

func reportFailureStage(
	output io.Writer,
	logger *slog.Logger,
	errorClass string,
	command string,
	startupStage string,
) {
	writeFailureLog(output, logger, errorClass, command, startupStage, "")
}

func writeFailureLog(
	output io.Writer,
	logger *slog.Logger,
	errorClass string,
	command string,
	startupStage string,
	errorDetail string,
) {
	if errorClass == "" {
		return
	}
	if logger == nil {
		if output == nil {
			return
		}
		logger = slog.New(slog.NewJSONHandler(output, &slog.HandlerOptions{Level: slog.LevelError}))
	}
	attributes := []any{
		slog.String("error_class", errorClass),
		slog.String("command", command),
	}
	if startupStage != "" {
		attributes = append(attributes, slog.String("startup_stage", startupStage))
	}
	if errorDetail != "" {
		attributes = append(attributes, slog.String("error_detail", errorDetail))
	}
	logger.Error("core migrator command failed", attributes...)
}

func reportDependencyStage(logger *slog.Logger, stage string) {
	if logger == nil || stage == "" {
		return
	}
	logger.Debug(
		"core migrator dependency preflight stage failed",
		slog.String("dependency_stage", stage),
	)
}

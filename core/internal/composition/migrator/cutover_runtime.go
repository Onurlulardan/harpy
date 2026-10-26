package migrator

import (
	"context"
	"io"
	"log/slog"

	"github.com/Onurlulardan/harpy/core/internal/platform/buildinfo"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/telemetry"
)

type cutoverRuntime struct {
	logger    *slog.Logger
	telemetry *telemetry.Runtime
	settings  config.Telemetry
}

func startCutoverRuntime(
	ctx context.Context,
	environment config.Environment,
	logLevel config.LogLevel,
	settings config.Telemetry,
	stderr io.Writer,
) (*cutoverRuntime, error) {
	info, err := buildinfo.ReadMigrator(environment)
	if err != nil {
		return nil, err
	}
	logger, err := telemetry.NewLogger(stderr, logLevel, info)
	if err != nil {
		return nil, err
	}
	runtime, err := telemetry.New(ctx, settings, info, logger)
	if err != nil {
		return nil, err
	}
	return &cutoverRuntime{logger: logger, telemetry: runtime, settings: settings}, nil
}

func (runtime *cutoverRuntime) shutdown() error {
	if runtime == nil || runtime.telemetry == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), runtime.settings.ExportTimeout())
	defer cancel()
	err := runtime.telemetry.Shutdown(ctx)
	runtime.telemetry = nil
	return err
}

package telemetry

import (
	"fmt"
	"io"
	"log"
	"log/slog"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/platform/buildinfo"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
)

func NewLogger(output io.Writer, level config.LogLevel, info buildinfo.Info) (*slog.Logger, error) {
	if output == nil {
		return nil, fmt.Errorf("telemetry log output is required")
	}
	leveler, err := slogLevel(level)
	if err != nil {
		return nil, err
	}
	handler := slog.NewJSONHandler(output, &slog.HandlerOptions{
		Level:       leveler,
		ReplaceAttr: redactAttribute,
	})
	version := info.Version()
	return slog.New(handler).With(
		"service.name", string(version.Service),
		"service.version", version.Version,
		"deployment.environment.name", string(version.Environment),
	), nil
}

func StandardErrorLog(logger *slog.Logger) (*log.Logger, error) {
	if logger == nil {
		return nil, fmt.Errorf("structured logger is required")
	}
	return slog.NewLogLogger(logger.Handler(), slog.LevelError), nil
}

func slogLevel(level config.LogLevel) (slog.Level, error) {
	switch level {
	case config.LogDebug:
		return slog.LevelDebug, nil
	case config.LogInfo:
		return slog.LevelInfo, nil
	case config.LogWarn:
		return slog.LevelWarn, nil
	case config.LogError:
		return slog.LevelError, nil
	default:
		return 0, fmt.Errorf("structured log level is invalid")
	}
}

func redactAttribute(groups []string, attribute slog.Attr) slog.Attr {
	qualified := strings.ToLower(strings.Join(append(groups, attribute.Key), "."))
	for _, marker := range []string{
		"authorization", "cookie", "credential", "database_url", "dsn",
		"password", "private_key", "secret", "token", "api_key", "apikey",
	} {
		if strings.Contains(qualified, marker) {
			return slog.String(attribute.Key, "[REDACTED]")
		}
	}
	return attribute
}

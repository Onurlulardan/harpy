package initialize

import (
	"context"
	"errors"
	"io"
	"os"

	"github.com/Onurlulardan/harpy/core/internal/platform/foundation"
	lifecycle "github.com/Onurlulardan/harpy/core/internal/platform/initialize"
)

func Run(arguments []string) int {
	return run(arguments, os.Stdout, os.Stderr, os.Getenv, os.ReadFile)
}

func run(
	arguments []string,
	stdout io.Writer,
	stderr io.Writer,
	lookup func(string) string,
	read func(string) ([]byte, error),
) int {
	if len(arguments) != 0 {
		_, _ = io.WriteString(stderr, "usage: core-initialize\n")
		return 2
	}
	cfg, err := loadConfig(lookup, read)
	if err != nil {
		_, _ = io.WriteString(stderr, foundation.Redact(err.Error(), nil)+"\n")
		return 2
	}
	result, err := lifecycle.Run(context.Background(), lifecycle.Request{
		Config: cfg, Goose: lifecycle.UnavailableGoose,
		Application: lifecycle.PermitApplication,
	})
	_, _ = io.WriteString(stdout, lifecycle.FormatReport(cfg.Database, result))
	if err == nil {
		return 0
	}
	_, _ = io.WriteString(stderr, err.Error()+"\n")
	var stage *lifecycle.StageError
	if errors.As(err, &stage) {
		return stage.ExitCode()
	}
	return 1
}

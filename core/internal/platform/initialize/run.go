package initialize

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/platform/foundation"
)

type Request struct {
	Config      foundation.Config
	Preflight   func(context.Context, foundation.Config) error
	Foundation  func(context.Context, foundation.Config) error
	Goose       func(context.Context) error
	Application func(context.Context) error
}

func Run(ctx context.Context, request Request) (Result, error) {
	var result Result
	if err := request.Config.Validate(); err != nil {
		result.fail(StageInitialize)
		return result, stageError(StageInitialize, "configuration", request.Config, false, 2, err)
	}
	if request.Goose == nil || request.Application == nil {
		result.fail(StageInitialize)
		return result, stageError(
			StageInitialize, "configuration", request.Config, false, 2,
			errors.New("goose and application stages are required"),
		)
	}
	preflight := request.Preflight
	if preflight == nil {
		preflight = foundation.Preflight
	}
	if err := preflight(ctx, request.Config); err != nil {
		result.fail(StageInitialize)
		return result, stageError(
			StageInitialize, "preflight", request.Config, retryableCause(err), 3, err,
		)
	}
	result.succeed(StageInitialize)
	apply := request.Foundation
	if apply == nil {
		apply = foundation.Apply
	}
	if err := apply(ctx, request.Config); err != nil {
		result.fail(StageFoundation)
		return result, stageError(
			StageFoundation, "foundation", request.Config, retryableCause(err), 4, err,
		)
	}
	result.succeed(StageFoundation)
	if err := request.Goose(ctx); err != nil {
		result.fail(StageGoose)
		return result, stageError(StageGoose, "goose", request.Config, true, 5, err)
	}
	result.succeed(StageGoose)
	if err := request.Application(ctx); err != nil {
		result.fail(StageApplication)
		return result, stageError(
			StageApplication, "application", request.Config, true, 6, err,
		)
	}
	result.succeed(StageApplication)
	return result, nil
}

func FormatReport(database string, result Result) string {
	var builder strings.Builder
	for _, stage := range result.Stages {
		status := "ok"
		if !stage.OK {
			status = "failed"
		}
		fmt.Fprintf(&builder, "stage=%s status=%s database=%s\n", stage.Name, status, database)
	}
	return builder.String()
}

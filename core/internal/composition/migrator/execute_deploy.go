package migrator

import (
	"context"
	"fmt"

	platformmigrator "github.com/Onurlulardan/harpy/core/internal/platform/migrator"
	"github.com/Onurlulardan/harpy/core/internal/platform/migratoroutput"
)

type deployRunner interface {
	Plan(context.Context) (platformmigrator.PlanResult, error)
	Verify(context.Context) (platformmigrator.VerifyResult, error)
	Up(context.Context, platformmigrator.UpRequest) (platformmigrator.UpResult, error)
}

type runIDFactory func() (string, error)

func executeDeploy(
	ctx context.Context,
	runner deployRunner,
	action deployAction,
	newRunID runIDFactory,
) (migratoroutput.Result, error) {
	if runner == nil || newRunID == nil {
		return migratoroutput.Result{}, deployRuntimeError(
			fmt.Errorf("deploy execution is incomplete"),
		)
	}
	switch action {
	case deployPlan:
		result, err := runner.Plan(ctx)
		if err != nil {
			return migratoroutput.Result{}, err
		}
		mapped, err := migratoroutput.FromPlan(result)
		if err != nil {
			return migratoroutput.Result{}, deployRuntimeError(err)
		}
		return mapped, nil
	case deployVerify:
		result, err := runner.Verify(ctx)
		if err != nil {
			return migratoroutput.Result{}, err
		}
		mapped, err := migratoroutput.FromVerify(result)
		if err != nil {
			return migratoroutput.Result{}, deployRuntimeError(err)
		}
		return mapped, nil
	case deployUp:
		runID, err := newRunID()
		if err != nil {
			return migratoroutput.Result{}, deployRuntimeError(
				fmt.Errorf("migration run ID generation failed"),
			)
		}
		result, err := runner.Up(ctx, platformmigrator.UpRequest{RunID: runID})
		if err != nil {
			return migratoroutput.Result{}, err
		}
		mapped, err := migratoroutput.FromUp(result)
		if err != nil {
			return migratoroutput.Result{}, deployRuntimeError(err)
		}
		return mapped, nil
	default:
		return migratoroutput.Result{}, deployRuntimeError(
			fmt.Errorf("deploy action is invalid"),
		)
	}
}

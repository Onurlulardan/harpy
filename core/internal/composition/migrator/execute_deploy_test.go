package migrator

import (
	"bytes"
	"context"
	"strings"
	"testing"

	platformmigrator "github.com/Onurlulardan/harpy/core/internal/platform/migrator"
)

type deployRunnerFixture struct{}

func (deployRunnerFixture) Plan(context.Context) (platformmigrator.PlanResult, error) {
	return platformmigrator.PlanResult{}, nil
}

func (deployRunnerFixture) Verify(context.Context) (platformmigrator.VerifyResult, error) {
	return platformmigrator.VerifyResult{
		DatabaseVersion: 1, DatabaseGeneration: 1, MinimumRuntimeGeneration: 1,
		VerifiedRunID: "11111111-1111-4111-8111-111111111111",
	}, nil
}

func (deployRunnerFixture) Up(
	_ context.Context,
	request platformmigrator.UpRequest,
) (platformmigrator.UpResult, error) {
	return platformmigrator.UpResult{
		RunID: request.RunID, AppliedVersions: []int64{1}, DatabaseVersion: 1,
	}, nil
}

func TestExecuteDeployWritesExactPlanResult(t *testing.T) {
	result, err := executeDeploy(
		context.Background(), deployRunnerFixture{}, deployPlan,
		func() (string, error) { return "11111111-1111-4111-8111-111111111111", nil },
	)
	if err != nil {
		t.Fatalf("execute plan: %v", err)
	}
	var output bytes.Buffer
	if err := result.Write(&output); err != nil {
		t.Fatalf("write plan: %v", err)
	}
	if output.Bytes()[output.Len()-1] != '\n' {
		t.Fatal("result has no trailing LF")
	}
	if !strings.Contains(output.String(), `"command":"plan"`) ||
		!strings.Contains(output.String(), `"pending":[]`) {
		t.Fatalf("unexpected plan output: %s", output.String())
	}
}

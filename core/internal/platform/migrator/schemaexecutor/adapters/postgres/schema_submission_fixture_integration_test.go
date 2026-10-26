//go:build integration

package postgres

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"testing"

	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/schemaexecutor"
	"github.com/jackc/pgx/v5/pgxpool"
)

const schemaSubmissionRuntimePassword = "schema-submission-runtime-only"

type schemaSubmissionInput struct {
	jobID    string
	plan     schemav1.Plan
	prepared schemav1.PreparedPlan
}

type schemaSubmissionReceipt struct {
	jobID, digest string
	generation    int64
	replay        bool
}

func schemaSubmissionUUID(value int) string {
	return fmt.Sprintf("018f47f4-7b1b-7cc0-8000-%012d", value)
}

func prepareSchemaSubmissionInput(
	t *testing.T,
	jobID string,
	plan schemav1.Plan,
) schemaSubmissionInput {
	t.Helper()
	prepared, err := schemav1.PreparePlan(plan)
	if err != nil {
		t.Fatalf("prepare schema submission: %v", err)
	}
	validated, err := schemaexecutor.BuildPlan(plan)
	if err != nil || validated.SHA256() != prepared.SHA256() ||
		!bytes.Equal(validated.CanonicalJSON(), prepared.CanonicalJSON()) {
		t.Fatalf("submission/executor preparation parity: %v", err)
	}
	validatedSteps, contractSteps := validated.Steps(), prepared.Steps()
	for index := range contractSteps {
		if validatedSteps[index].SHA256() != contractSteps[index].SHA256() {
			t.Fatalf("step %d preparation digest differs", index)
		}
	}
	return schemaSubmissionInput{jobID: jobID, plan: plan, prepared: prepared}
}

func (input schemaSubmissionInput) stepHex() []string {
	steps := input.prepared.Steps()
	encoded := make([]string, len(steps))
	for index := range steps {
		encoded[index] = hex.EncodeToString(steps[index].CanonicalPreimage())
	}
	return encoded
}

func callSchemaSubmission(
	ctx context.Context,
	pool *pgxpool.Pool,
	input schemaSubmissionInput,
	planBytes []byte,
	digest string,
	steps []string,
) (schemaSubmissionReceipt, error) {
	encodedSteps, err := json.Marshal(steps)
	if err != nil {
		return schemaSubmissionReceipt{}, err
	}
	var receipt schemaSubmissionReceipt
	err = pool.QueryRow(ctx, `SELECT job_id::text, generation, plan_sha256, replay
FROM forms_submit_schema_plan_v1($1,$2,$3,$4,$5,$6,$7,$8,$9,$10::jsonb)`,
		input.jobID, string(input.plan.Resource.Kind), input.plan.Resource.ID,
		input.plan.Resource.Revision, input.plan.ContractVersion,
		input.plan.Compiler.Version, input.plan.Compiler.PostgreSQLMajor,
		planBytes, digest, string(encodedSteps),
	).Scan(&receipt.jobID, &receipt.generation, &receipt.digest, &receipt.replay)
	return receipt, err
}

func submitSchemaInput(
	ctx context.Context,
	pool *pgxpool.Pool,
	input schemaSubmissionInput,
) (schemaSubmissionReceipt, error) {
	return callSchemaSubmission(ctx, pool, input, input.prepared.CanonicalJSON(),
		input.prepared.SHA256(), input.stepHex())
}

func openSchemaSubmissionRuntime(
	t *testing.T,
	ctx context.Context,
	environment *schemaExecutorIntegrationEnvironment,
	schemaName string,
) *pgxpool.Pool {
	t.Helper()
	if _, err := environment.admin.Exec(ctx, `ALTER ROLE harpy_v2_runtime PASSWORD '`+
		schemaSubmissionRuntimePassword+`'`); err != nil {
		t.Fatal(err)
	}
	config, err := pgxpool.ParseConfig(environment.dsn)
	if err != nil {
		t.Fatal(err)
	}
	config.ConnConfig.Database = "harpy_adapter"
	config.ConnConfig.User = "harpy_v2_runtime"
	config.ConnConfig.Password = schemaSubmissionRuntimePassword
	config.ConnConfig.RuntimeParams = map[string]string{
		"search_path": schemaName + ",pg_catalog",
	}
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	return pool
}

package runtimev1

import (
	"encoding/json"
	"testing"
)

const cutoverResultDigest = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func TestCutoverCommandResultBranchesMatchPublishedSchema(t *testing.T) {
	tests := []struct {
		name   string
		result CutoverCommandResult
		code   ExitCode
	}{
		{"adoption", mustAdoptionResult(t), ExitCodeClean},
		{"completed", mustFinalizationResult(t, "completed"), ExitCodeClean},
		{"failed", mustFinalizationResult(t, "failed"), ExitCodeDeployMigration},
		{"cancelled", mustFinalizationResult(t, "cancelled"), ExitCodeDeployMigration},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := test.result.Validate(); err != nil || test.result.ExitCode != test.code {
				t.Fatalf("result validation/code = %v/%d", err, test.result.ExitCode)
			}
			document, err := json.Marshal(test.result)
			if err != nil {
				t.Fatal(err)
			}
			var decoded any
			if err := json.Unmarshal(document, &decoded); err != nil {
				t.Fatal(err)
			}
			if err := compileSchema(t, "cutover-command-result.schema.json").Validate(decoded); err != nil {
				t.Fatalf("result violates published schema: %v", err)
			}
		})
	}
}

func TestCutoverCommandResultRejectsFalseSuccessAndMixedFields(t *testing.T) {
	failed := mustFinalizationResult(t, "failed")
	failed.Status = "succeeded"
	failed.ExitCode = ExitCodeClean
	if err := failed.Validate(); err == nil {
		t.Fatal("failed terminal reported as clean success")
	}
	completed := mustFinalizationResult(t, "completed")
	completed.BatchSHA256 = cutoverString(cutoverResultDigest)
	if err := completed.Validate(); err == nil {
		t.Fatal("mixed cutover result branch accepted")
	}
}

func TestCutoverCommandResultSchemaRejectsFalseSuccess(t *testing.T) {
	result := mustFinalizationResult(t, "failed")
	document, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(document, &decoded); err != nil {
		t.Fatal(err)
	}
	decoded["status"] = "succeeded"
	decoded["exit_code"] = float64(0)
	if err := compileSchema(t, "cutover-command-result.schema.json").Validate(decoded); err == nil {
		t.Fatal("published schema accepted failed terminal as clean success")
	}
}

func TestDecodeCutoverCommandResultRejectsUnknownAndTrailingContent(t *testing.T) {
	result := mustFinalizationResult(t, "completed")
	document, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := DecodeCutoverCommandResult(document); err != nil {
		t.Fatalf("valid result decode failed: %v", err)
	}
	unknown := append(document[:len(document)-1], []byte(`,"unknown":true}`)...)
	if _, err := DecodeCutoverCommandResult(unknown); err == nil {
		t.Fatal("unknown output field accepted")
	}
	if _, err := DecodeCutoverCommandResult(append(document, document...)); err == nil {
		t.Fatal("trailing output object accepted")
	}
}

func mustAdoptionResult(t *testing.T) CutoverCommandResult {
	t.Helper()
	result, err := NewCutoverAdoptionCommandResult(
		"11111111-1111-4111-8111-111111111111",
		"22222222-2222-4222-8222-222222222222",
		cutoverResultDigest, 1, cutoverResultDigest, cutoverResultDigest,
	)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func mustFinalizationResult(t *testing.T, state string) CutoverCommandResult {
	t.Helper()
	result, err := NewCutoverFinalizationCommandResult(
		"11111111-1111-4111-8111-111111111111",
		"22222222-2222-4222-8222-222222222222", state, cutoverResultDigest,
	)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

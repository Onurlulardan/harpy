package migrator

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/processcode"
)

const commandResultDigest = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func TestWriteFinalizationResultMapsTerminalStateToProcessCode(t *testing.T) {
	tests := []struct {
		state  string
		status string
		code   processcode.Code
	}{
		{"completed", "succeeded", processcode.Clean},
		{"failed", "failed", processcode.DeployMigration},
		{"cancelled", "cancelled", processcode.DeployMigration},
	}
	for _, test := range tests {
		t.Run(test.state, func(t *testing.T) {
			var output bytes.Buffer
			code, err := writeFinalizationResult(
				&output, "11111111-1111-4111-8111-111111111111",
				"22222222-2222-4222-8222-222222222222", test.state,
				commandResultDigest,
			)
			if err != nil || code != test.code {
				t.Fatalf("write result/code = %v/%d", err, code)
			}
			var result struct {
				Status   string `json:"status"`
				ExitCode int    `json:"exit_code"`
			}
			if err := json.Unmarshal(output.Bytes(), &result); err != nil {
				t.Fatal(err)
			}
			if result.Status != test.status || result.ExitCode != test.code.Int() {
				t.Fatalf("decoded result = %#v", result)
			}
		})
	}
}

func TestWriteAdoptionResultUsesTypedSuccessContract(t *testing.T) {
	var output bytes.Buffer
	err := writeAdoptionResult(
		&output, "11111111-1111-4111-8111-111111111111",
		"22222222-2222-4222-8222-222222222222", commandResultDigest,
		0, commandResultDigest, commandResultDigest,
	)
	var result struct {
		Status   string `json:"status"`
		ExitCode int    `json:"exit_code"`
	}
	decodeErr := json.Unmarshal(output.Bytes(), &result)
	if err != nil || decodeErr != nil || result.Status != "succeeded" ||
		result.ExitCode != processcode.Clean.Int() {
		t.Fatalf("adoption result = %#v write=%v decode=%v", result, err, decodeErr)
	}
}

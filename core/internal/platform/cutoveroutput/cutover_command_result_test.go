package cutoveroutput

import (
	"bytes"
	"testing"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
)

const resultDigest = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func TestResultWritesValidatedContract(t *testing.T) {
	result, err := FromFinalization(
		"11111111-1111-4111-8111-111111111111",
		"22222222-2222-4222-8222-222222222222", "failed", resultDigest,
	)
	if err != nil || result.ExitCode() != runtimev1.ExitCodeDeployMigration {
		t.Fatalf("result/code = %v/%d", err, result.ExitCode())
	}
	var output bytes.Buffer
	if err := result.Encode(&output); err != nil {
		t.Fatal(err)
	}
	decoded, err := runtimev1.DecodeCutoverCommandResult(output.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if err := decoded.Validate(); err != nil {
		t.Fatalf("written contract is invalid: %v", err)
	}
}

func TestResultRejectsInvalidAdoption(t *testing.T) {
	_, err := FromAdoption(
		"11111111-1111-4111-8111-111111111111",
		"22222222-2222-4222-8222-222222222222", resultDigest,
		-1, resultDigest, resultDigest,
	)
	if err == nil {
		t.Fatal("negative adoption result accepted")
	}
}

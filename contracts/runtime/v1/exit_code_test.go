package runtimev1

import "testing"

func TestExitCodesMatchProcessContract(t *testing.T) {
	t.Parallel()
	tests := []struct {
		code ExitCode
		want int
	}{
		{code: ExitCodeClean, want: 0},
		{code: ExitCodeInvalidStartup, want: 2},
		{code: ExitCodeDependencyPreflight, want: 3},
		{code: ExitCodeServeFailure, want: 4},
		{code: ExitCodeDrainForced, want: 5},
		{code: ExitCodeDeployMigration, want: 10},
	}
	for _, test := range tests {
		if !test.code.Valid() {
			t.Errorf("expected exit code %d to be valid", test.code)
		}
		if test.code.Int() != test.want {
			t.Errorf("exit code %d: want %d", test.code, test.want)
		}
	}
}

func TestExitCodeRejectsUnknownValue(t *testing.T) {
	t.Parallel()
	if ExitCode(1).Valid() {
		t.Fatal("unexpected exit code accepted")
	}
}

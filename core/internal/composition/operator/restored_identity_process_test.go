package operator

import (
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/processcode"
)

type identitySQLStateError string

func (value identitySQLStateError) Error() string    { return string(value) }
func (value identitySQLStateError) SQLState() string { return string(value) }

func TestIdentityReconcileExitClassesAreExact(t *testing.T) {
	tests := []struct {
		stage string
		err   error
		want  processcode.Code
	}{
		{"configuration", errors.New("bad config"), processcode.InvalidStartup},
		{"build_contract", errors.New("bad build"), processcode.InvalidStartup},
		{"database", errors.New("offline"), processcode.DependencyPreflight},
		{"composition", errors.New("binding"), processcode.DependencyPreflight},
		{"evidence", errors.New("tamper"), processcode.OperatorTerminal},
		{"custody", errors.New("not staged"), processcode.OperatorResumable},
		{"reconcile", identitySQLStateError("42501"), processcode.OperatorTerminal},
		{"reconcile", identitySQLStateError("08006"), processcode.OperatorResumable},
	}
	for _, test := range tests {
		if got := classifyIdentityReconcileError(test.stage, test.err); got != test.want {
			t.Fatalf("stage=%s code=%d want=%d", test.stage, got, test.want)
		}
	}
}

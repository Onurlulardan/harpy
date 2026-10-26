package operator

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/orchestration/environmentoperator"
	"github.com/Onurlulardan/harpy/core/internal/platform/processcode"
)

func TestRunRejectsEveryNonInitializeInvocationWithoutStdout(t *testing.T) {
	for _, arguments := range [][]string{nil, {}, {"status"}, {"initialize", "extra"}} {
		var stdout, stderr bytes.Buffer
		code := run(context.Background(), arguments, &stdout, &stderr)
		if code != processcode.InvalidStartup || stdout.Len() != 0 ||
			stderr.String() != "{\"success\":false,\"stage\":\"invocation\",\"exit_code\":2}\n" {
			t.Fatalf("arguments=%v code=%d stdout=%q stderr=%q",
				arguments, code, stdout.String(), stderr.String())
		}
	}
}

func TestInitializeExitClassesAreExact(t *testing.T) {
	tests := []struct {
		err  error
		want processcode.Code
	}{
		{environmentoperator.ErrInvalid, processcode.InvalidStartup},
		{environmentoperator.ErrDependencyMismatch, processcode.DependencyPreflight},
		{environmentoperator.ErrResumable, processcode.OperatorResumable},
		{environmentoperator.ErrFailed, processcode.OperatorTerminal},
		{environmentoperator.ErrFenced, processcode.OperatorTerminal},
		{errors.New("transient"), processcode.OperatorResumable},
	}
	for _, test := range tests {
		if got := classifyInitializeError(test.err); got != test.want {
			t.Fatalf("error=%v code=%d want=%d", test.err, got, test.want)
		}
	}
}

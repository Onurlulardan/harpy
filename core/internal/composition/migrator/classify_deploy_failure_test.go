package migrator

import (
	"errors"
	"testing"

	platformmigrator "github.com/Onurlulardan/harpy/core/internal/platform/migrator"
	"github.com/Onurlulardan/harpy/core/internal/platform/processcode"
)

func TestClassifyDeployFailureUsesContractedExitCodes(t *testing.T) {
	tests := []struct {
		err       error
		wantCode  processcode.Code
		wantClass string
	}{
		{deployRuntimeError(errors.New("write")), processcode.ServeFailure, "runtime"},
		{platformmigrator.ErrDependencyPreflight, processcode.DependencyPreflight, "dependency_preflight"},
		{platformmigrator.ErrMigrationsPending, processcode.DeployMigration, "migration"},
	}
	for _, test := range tests {
		code, class := classifyDeployFailure(test.err)
		if code != test.wantCode || class != test.wantClass {
			t.Fatalf("classification=(%d,%q), want (%d,%q)", code, class, test.wantCode, test.wantClass)
		}
	}
}

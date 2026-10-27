package execution

import (
	"errors"
	"testing"

	restexecution "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/execution"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type closedDriverPort struct{}

func (*closedDriverPort) Execute(transaction.TenantScope, restexecution.Program) error {
	return nil
}

func TestExecuteRequiresLiveSnapshotScopeAndProgram(t *testing.T) {
	executor, err := NewExecutor(&closedDriverPort{})
	if err != nil {
		t.Fatal(err)
	}
	if err := executor.Execute(transaction.TenantScope{}, restexecution.Program{}); !errors.Is(
		err, ErrContract,
	) {
		t.Fatalf("invalid execution error = %v", err)
	}
}

func TestNewExecutorRejectsTypedNilDriver(t *testing.T) {
	var driver *closedDriverPort
	if _, err := NewExecutor(driver); !errors.Is(err, ErrContract) {
		t.Fatalf("typed nil error = %v", err)
	}
}

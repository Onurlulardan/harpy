package executionpostgres

import (
	"testing"

	restexecution "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/execution"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestExecuteRejectsClosedScopeBeforeBinding(t *testing.T) {
	port, err := New()
	if err != nil {
		t.Fatal(err)
	}
	if err := port.Execute(transaction.TenantScope{}, restexecution.Program{}); err == nil {
		t.Fatal("closed scope was accepted")
	}
}

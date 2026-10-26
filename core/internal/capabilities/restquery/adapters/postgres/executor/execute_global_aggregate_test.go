package executor

import (
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestExecuteGlobalAggregateRejectsUnconfiguredExecutor(t *testing.T) {
	_, err := (GlobalExecutor{}).ExecuteAggregate(
		t.Context(), platformtx.GlobalScope{}, resolved.AuthorizedPlan{},
	)
	if !errors.Is(err, ErrContract) {
		t.Fatalf("ExecuteAggregate error = %v, want ErrContract", err)
	}
}

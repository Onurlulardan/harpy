//go:build integration

package executor

import (
	"context"
	"testing"

	v1query "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (gate *pg18Gate) pg18ExecuteV1(
	t *testing.T,
	executor Executor,
	tenant string,
	accepted v1query.Accepted,
) Result {
	t.Helper()
	var result Result
	err := gate.coordinator.InTenant(
		gate.ctx, pg18ScopeRequest(t, tenant, "forms"),
		func(ctx context.Context, scope transaction.TenantScope) error {
			var err error
			result, err = executor.ExecuteV1(ctx, scope, accepted)
			return err
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

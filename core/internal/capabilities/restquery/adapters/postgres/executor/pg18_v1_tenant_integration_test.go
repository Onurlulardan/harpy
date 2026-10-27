//go:build integration

package executor

import (
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"
)

func (gate *pg18Gate) verifyV1TenantIsolation(t *testing.T) {
	port := pg18NewOwnerPort(t, "forms")
	executor := pg18NewExecutor(t, gate.codec, planner.HardPolicy(), port)
	accepted := pg18V1Plan(t, "page=2&limit=1", pg18TenantB)
	result := gate.pg18ExecuteV1(t, executor, pg18TenantB, accepted)
	identifiers := pg18ResultIDs(t, result)
	if len(identifiers) != 1 ||
		identifiers[0] != "50000000-0000-4000-8000-000000000002" ||
		strings.HasPrefix(identifiers[0], "40000000") {
		t.Fatalf("tenant B page leaked: %v", identifiers)
	}
	pg18AssertExactTotal(t, result, 2)
}

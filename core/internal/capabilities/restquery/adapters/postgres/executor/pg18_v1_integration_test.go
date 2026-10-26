//go:build integration

package executor

import "testing"

func (gate *pg18Gate) verifyV1Execution(t *testing.T) {
	gate.seedV1StableRows(t)
	t.Run("page2 page3 stable tie and offset oracle", gate.verifyV1Pages)
	t.Run("missing predecessor is empty with exact count", gate.verifyV1Missing)
	t.Run("tenant isolation", gate.verifyV1TenantIsolation)
	t.Run("injection remains bound", gate.verifyV1Injection)
}
